package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func GenerateCodeVerifier() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func CodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

type oidcConfig struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
}

func discoverEndpoints(issuer string) (authURL, tokenURL string) {
	fallbackAuth := strings.TrimRight(issuer, "/") + "/v1/authorize"
	fallbackToken := strings.TrimRight(issuer, "/") + "/v1/token"

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration")
	if err != nil {
		return fallbackAuth, fallbackToken
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fallbackAuth, fallbackToken
	}

	var cfg oidcConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return fallbackAuth, fallbackToken
	}

	if cfg.AuthorizationEndpoint == "" {
		cfg.AuthorizationEndpoint = fallbackAuth
	}
	if cfg.TokenEndpoint == "" {
		cfg.TokenEndpoint = fallbackToken
	}
	return cfg.AuthorizationEndpoint, cfg.TokenEndpoint
}

func randomState() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func pickPort() (int, net.Listener, error) {
	offset, err := rand.Int(rand.Reader, big.NewInt(24))
	if err != nil {
		return 0, nil, err
	}
	start := 9876 + int(offset.Int64())

	for i := 0; i < 24; i++ {
		port := 9876 + (start-9876+i)%24
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return port, ln, nil
		}
	}
	return 0, nil, fmt.Errorf("no available port in range 9876-9899")
}

func StartPKCEFlow(issuer, clientID string) (token string, email string, err error) {
	verifier := GenerateCodeVerifier()
	challenge := CodeChallenge(verifier)
	state := randomState()

	authEndpoint, tokenEndpoint := discoverEndpoints(issuer)

	port, ln, err := pickPort()
	if err != nil {
		return "", "", fmt.Errorf("picking callback port: %w", err)
	}

	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	params := url.Values{
		"client_id":             {clientID},
		"response_type":        {"code"},
		"scope":                {"openid email profile"},
		"redirect_uri":         {redirectURI},
		"code_challenge":       {challenge},
		"code_challenge_method": {"S256"},
		"state":                {state},
	}
	authURL := authEndpoint + "?" + params.Encode()

	type callbackResult struct {
		code string
		err  error
	}
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			resultCh <- callbackResult{err: fmt.Errorf("state mismatch: expected %s, got %s", state, q.Get("state"))}
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		if errParam := q.Get("error"); errParam != "" {
			desc := q.Get("error_description")
			resultCh <- callbackResult{err: fmt.Errorf("authorization error: %s: %s", errParam, desc)}
			http.Error(w, errParam+": "+desc, http.StatusBadRequest)
			return
		}

		code := q.Get("code")
		if code == "" {
			resultCh <- callbackResult{err: fmt.Errorf("no authorization code in callback")}
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html><html><body><h2>Login successful! You can close this tab.</h2></body></html>`)
		resultCh <- callbackResult{code: code}
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)

	if err := openBrowser(authURL); err != nil {
		// Non-fatal; user can copy the URL.
	}
	fmt.Printf("Opening browser for login... If it doesn't open, visit:\n%s\n", authURL)

	select {
	case res := <-resultCh:
		if res.err != nil {
			srv.Close()
			return "", "", res.err
		}

		tok, em, err := exchangeCode(tokenEndpoint, res.code, redirectURI, clientID, verifier)
		srv.Close()
		if err != nil {
			return "", "", err
		}
		return tok, em, nil

	case <-time.After(2 * time.Minute):
		srv.Close()
		return "", "", fmt.Errorf("login timed out after 2 minutes")
	}
}

func exchangeCode(tokenEndpoint, code, redirectURI, clientID, verifier string) (string, string, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"code_verifier": {verifier},
	}

	resp, err := http.PostForm(tokenEndpoint, form)
	if err != nil {
		return "", "", fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, body)
	}

	var tokenResp struct {
		IDToken     string `json:"id_token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", "", fmt.Errorf("parsing token response: %w", err)
	}

	tok := tokenResp.IDToken
	if tok == "" {
		tok = tokenResp.AccessToken
	}
	if tok == "" {
		return "", "", fmt.Errorf("no id_token or access_token in response: %s", body)
	}

	email, err := extractEmailFromJWT(tok)
	if err != nil {
		return tok, "", fmt.Errorf("extracting email from token: %w", err)
	}

	return tok, email, nil
}

func extractEmailFromJWT(tok string) (string, error) {
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("token has %d segments, expected 3", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decoding token payload: %w", err)
	}

	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Sub               string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("parsing token claims: %w", err)
	}

	switch {
	case claims.Email != "":
		return claims.Email, nil
	case claims.PreferredUsername != "":
		return claims.PreferredUsername, nil
	case claims.Sub != "":
		return claims.Sub, nil
	default:
		return "", fmt.Errorf("no email, preferred_username, or sub claim in token")
	}
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
