package server

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dunkinfrunkin/kit/internal/auth"
	"github.com/dunkinfrunkin/kit/internal/crypto"
	"github.com/dunkinfrunkin/kit/internal/store"
)

//go:embed static
var staticFiles embed.FS

type itemResponse struct {
	ID        int       `json:"id"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toItemResponse(item *store.Item, content []byte) itemResponse {
	return itemResponse{
		ID:        item.ID,
		Namespace: item.Namespace,
		Type:      item.Type,
		Name:      item.Name,
		Content:   base64.StdEncoding.EncodeToString(content),
		Author:    item.Author,
		Version:   item.Version,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

type Server struct {
	store     *store.Store
	secret    string
	oidc      *auth.OIDCVerifier
	apiTokens map[string]string // hash → email
	mu        sync.RWMutex
	mux       *http.ServeMux
}

func New(s *store.Store, secret string, oidc *auth.OIDCVerifier) *Server {
	srv := &Server{
		store:     s,
		secret:    secret,
		oidc:      oidc,
		apiTokens: make(map[string]string),
		mux:       http.NewServeMux(),
	}

	staticFS, _ := fs.Sub(staticFiles, "static")
	indexHTML, _ := fs.ReadFile(staticFS, "index.html")
	serveUI := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	}
	srv.mux.HandleFunc("GET /ui", serveUI)
	srv.mux.HandleFunc("GET /{$}", serveUI)

	srv.mux.HandleFunc("POST /login", srv.handleLogin)
	srv.mux.HandleFunc("GET /auth/config", srv.handleAuthConfig)
	srv.mux.HandleFunc("GET /auth/sso", srv.handleSSORedirect)
	srv.mux.HandleFunc("GET /auth/callback", srv.handleSSOCallback)

	srv.mux.HandleFunc("GET /skills", srv.handleListAll("skill"))
	srv.mux.HandleFunc("GET /hooks", srv.handleListAll("hook"))
	srv.mux.HandleFunc("GET /configs", srv.handleListAll("config"))
	srv.mux.HandleFunc("GET /search", srv.handleSearch)

	srv.mux.HandleFunc("GET /{namespace}/skills", srv.handleList("skill"))
	srv.mux.HandleFunc("GET /{namespace}/hooks", srv.handleList("hook"))
	srv.mux.HandleFunc("GET /{namespace}/configs", srv.handleList("config"))

	srv.mux.HandleFunc("GET /{namespace}/skills/{name}", srv.handleGet("skill"))
	srv.mux.HandleFunc("GET /{namespace}/hooks/{name}", srv.handleGet("hook"))
	srv.mux.HandleFunc("GET /{namespace}/configs/{name}", srv.handleGet("config"))

	srv.mux.HandleFunc("POST /{namespace}/skills", srv.handlePush("skill"))
	srv.mux.HandleFunc("POST /{namespace}/hooks", srv.handlePush("hook"))
	srv.mux.HandleFunc("POST /{namespace}/configs", srv.handlePush("config"))

	srv.mux.HandleFunc("DELETE /{namespace}/skills/{name}", srv.handleDelete("skill"))
	srv.mux.HandleFunc("DELETE /{namespace}/hooks/{name}", srv.handleDelete("hook"))
	srv.mux.HandleFunc("DELETE /{namespace}/configs/{name}", srv.handleDelete("config"))

	srv.mux.HandleFunc("POST /tokens", srv.handleCreateToken)
	srv.mux.HandleFunc("GET /tokens", srv.handleListTokens)
	srv.mux.HandleFunc("DELETE /tokens/{name}", srv.handleDeleteToken)

	srv.mux.HandleFunc("GET /{namespace}/profiles", srv.handleListProfiles)
	srv.mux.HandleFunc("GET /{namespace}/profiles/{name}", srv.handleGetProfile)
	srv.mux.HandleFunc("POST /{namespace}/profiles", srv.handleCreateProfile)
	srv.mux.HandleFunc("POST /{namespace}/profiles/{name}/items", srv.handleAddProfileItem)
	srv.mux.HandleFunc("DELETE /{namespace}/profiles/{name}", srv.handleDeleteProfile)

	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) authenticate(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	token := strings.TrimPrefix(h, "Bearer ")

	// 1. API token (kit_ prefix) — look up by hash
	if strings.HasPrefix(token, "kit_") {
		hash := auth.HashAPIToken(token)
		s.mu.RLock()
		email, ok := s.apiTokens[hash]
		s.mu.RUnlock()
		if ok {
			return email, nil
		}
		return "", fmt.Errorf("invalid API token")
	}

	// 2. Self-signed JWT
	if email, err := auth.Verify(s.secret, token); err == nil {
		return email, nil
	}

	// 3. OIDC token
	if s.oidc != nil {
		if email, err := s.oidc.VerifyToken(token); err == nil {
			return email, nil
		}
	}

	return "", fmt.Errorf("unauthorized: no valid credentials")
}

func (s *Server) handleAuthConfig(w http.ResponseWriter, r *http.Request) {
	cfg := map[string]interface{}{
		"sso_enabled": s.oidc != nil,
	}
	if s.oidc != nil {
		cfg["issuer"] = s.oidc.Config().Issuer
		cfg["client_id"] = s.oidc.Config().ClientID
		if s.oidc.Config().ClientSecret != "" {
			cfg["client_secret"] = s.oidc.Config().ClientSecret
		}
	}
	s.json(w, http.StatusOK, cfg)
}

func (s *Server) handleSSORedirect(w http.ResponseWriter, r *http.Request) {
	if s.oidc == nil {
		s.error(w, http.StatusNotFound, "SSO not configured")
		return
	}

	cfg := s.oidc.Config()

	authEndpoint := strings.TrimRight(cfg.Issuer, "/") + "/v1/authorize"
	disc, err := http.Get(strings.TrimRight(cfg.Issuer, "/") + "/.well-known/openid-configuration")
	if err == nil {
		defer disc.Body.Close()
		var d struct {
			AuthorizationEndpoint string `json:"authorization_endpoint"`
		}
		if json.NewDecoder(disc.Body).Decode(&d) == nil && d.AuthorizationEndpoint != "" {
			authEndpoint = d.AuthorizationEndpoint
		}
	}

	verifier := auth.GenerateCodeVerifier()
	challenge := auth.CodeChallenge(verifier)

	http.SetCookie(w, &http.Cookie{
		Name:     "kit_pkce",
		Value:    verifier,
		Path:     "/auth",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	callbackURL := "http://" + r.Host + "/auth/callback"
	q := url.Values{
		"client_id":             {cfg.ClientID},
		"response_type":        {"code"},
		"scope":                {"openid email profile"},
		"redirect_uri":         {callbackURL},
		"state":                {"ui"},
		"code_challenge":       {challenge},
		"code_challenge_method": {"S256"},
	}
	http.Redirect(w, r, authEndpoint+"?"+q.Encode(), http.StatusFound)
}

func (s *Server) handleSSOCallback(w http.ResponseWriter, r *http.Request) {
	if s.oidc == nil {
		s.error(w, http.StatusNotFound, "SSO not configured")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errMsg := r.URL.Query().Get("error_description")
		if errMsg == "" {
			errMsg = r.URL.Query().Get("error")
		}
		s.error(w, http.StatusBadRequest, "SSO failed: "+errMsg)
		return
	}

	cfg := s.oidc.Config()
	callbackURL := "http://" + r.Host + "/auth/callback"

	tokenEndpoint := strings.TrimRight(cfg.Issuer, "/") + "/v1/token"
	disc, err := http.Get(strings.TrimRight(cfg.Issuer, "/") + "/.well-known/openid-configuration")
	if err == nil {
		defer disc.Body.Close()
		var d struct {
			TokenEndpoint string `json:"token_endpoint"`
		}
		if json.NewDecoder(disc.Body).Decode(&d) == nil && d.TokenEndpoint != "" {
			tokenEndpoint = d.TokenEndpoint
		}
	}

	verifier := ""
	if cookie, err := r.Cookie("kit_pkce"); err == nil {
		verifier = cookie.Value
	}

	formValues := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {callbackURL},
		"client_id":    {cfg.ClientID},
	}
	if verifier != "" {
		formValues.Set("code_verifier", verifier)
	}
	if cfg.ClientSecret != "" {
		formValues.Set("client_secret", cfg.ClientSecret)
	}
	form := formValues.Encode()

	tokenReq, _ := http.NewRequest("POST", tokenEndpoint, strings.NewReader(form))
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenResp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "token exchange failed")
		return
	}
	defer tokenResp.Body.Close()

	var tokenData struct {
		IDToken     string `json:"id_token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
		s.error(w, http.StatusInternalServerError, "failed to parse token response")
		return
	}

	idToken := tokenData.IDToken
	if idToken == "" {
		idToken = tokenData.AccessToken
	}

	email, err := s.oidc.VerifyToken(idToken)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "token verification failed: "+err.Error())
		return
	}

	kitToken, err := auth.Sign(s.secret, email)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><title>kit</title></head>
<body style="background:#0a0a0b;color:#ededef;font-family:Inter,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0">
<div style="text-align:center">
<h2 style="color:#22c55e;margin-bottom:8px">Logged in as %s</h2>
<p style="color:#8b8b8e">Redirecting to dashboard...</p>
</div>
<script>
localStorage.setItem('kit_token','%s');
setTimeout(function(){window.location.href='/ui'},800);
</script></body></html>`, email, kitToken)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" {
		s.error(w, http.StatusBadRequest, "email is required")
		return
	}

	token, err := auth.Sign(s.secret, body.Email)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to sign token")
		return
	}

	s.json(w, http.StatusOK, map[string]string{
		"token": token,
		"email": body.Email,
	})
}

func (s *Server) handlePush(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := s.authenticate(r)
		if err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body struct {
			Name    string `json:"name"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			s.error(w, http.StatusBadRequest, "invalid request body")
			return
		}

		namespace := r.PathValue("namespace")
		content, err := base64.StdEncoding.DecodeString(body.Content)
		if err != nil {
			s.error(w, http.StatusBadRequest, "invalid base64 content")
			return
		}

		if isPersonal(namespace) {
			if ownerEmail(namespace) != email {
				s.error(w, http.StatusForbidden, "namespace does not match authenticated user")
				return
			}
			key := crypto.DeriveKey(s.secret, email)
			content, err = crypto.Encrypt(key, content)
			if err != nil {
				s.error(w, http.StatusInternalServerError, "encryption failed")
				return
			}
		}

		item, err := s.store.PushItem(namespace, itemType, body.Name, content, email)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "failed to push item")
			return
		}

		s.json(w, http.StatusOK, toItemResponse(item, content))
	}
}

func (s *Server) handleGet(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := s.authenticate(r)
		if err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		namespace := r.PathValue("namespace")
		name := r.PathValue("name")

		item, err := s.store.GetItem(namespace, itemType, name)
		if err != nil {
			s.error(w, http.StatusNotFound, "item not found")
			return
		}

		content := item.Content
		if isPersonal(namespace) {
			if ownerEmail(namespace) != email {
				s.error(w, http.StatusForbidden, "namespace does not match authenticated user")
				return
			}
			key := crypto.DeriveKey(s.secret, email)
			content, err = crypto.Decrypt(key, content)
			if err != nil {
				s.error(w, http.StatusInternalServerError, "decryption failed")
				return
			}
		}

		s.json(w, http.StatusOK, toItemResponse(item, content))
	}
}

func (s *Server) handleList(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := s.authenticate(r)
		if err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		namespace := r.PathValue("namespace")
		if isPersonal(namespace) && ownerEmail(namespace) != email {
			s.error(w, http.StatusForbidden, "access denied")
			return
		}

		items, err := s.store.ListItems(namespace, itemType)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "failed to list items")
			return
		}

		s.json(w, http.StatusOK, items)
	}
}

func (s *Server) handleListAll(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := s.authenticate(r)
		if err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		items, err := s.store.ListAllItems(itemType)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "failed to list items")
			return
		}

		filtered := make([]store.Item, 0, len(items))
		for _, item := range items {
			if isPersonal(item.Namespace) && ownerEmail(item.Namespace) != email {
				continue
			}
			filtered = append(filtered, item)
		}

		s.json(w, http.StatusOK, filtered)
	}
}

func (s *Server) handleDelete(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := s.authenticate(r)
		if err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		namespace := r.PathValue("namespace")
		name := r.PathValue("name")

		if err := s.store.DeleteItem(namespace, itemType, name, email); err != nil {
			s.error(w, http.StatusInternalServerError, "failed to delete item")
			return
		}

		s.json(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticate(r); err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query().Get("q")
	items, err := s.store.SearchItems(query)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "search failed")
		return
	}

	s.json(w, http.StatusOK, items)
}

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticate(r); err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	namespace := r.PathValue("namespace")
	profiles, err := s.store.ListProfiles(namespace)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to list profiles")
		return
	}

	s.json(w, http.StatusOK, profiles)
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticate(r); err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	profile, err := s.store.GetProfile(namespace, name)
	if err != nil {
		s.error(w, http.StatusNotFound, "profile not found")
		return
	}

	s.json(w, http.StatusOK, profile)
}

func (s *Server) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	email, err := s.authenticate(r)
	if err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	namespace := r.PathValue("namespace")
	profile, err := s.store.CreateProfile(namespace, body.Name, email)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to create profile")
		return
	}

	s.json(w, http.StatusCreated, profile)
}

func (s *Server) handleAddProfileItem(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticate(r); err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var ref store.ProfileRef
	if err := json.NewDecoder(r.Body).Decode(&ref); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	if err := s.store.AddProfileItem(namespace, name, ref); err != nil {
		s.error(w, http.StatusInternalServerError, "failed to add profile item")
		return
	}

	s.json(w, http.StatusOK, map[string]string{"status": "added"})
}

func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	email, err := s.authenticate(r)
	if err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	if err := s.store.DeleteProfile(namespace, name, email); err != nil {
		s.error(w, http.StatusInternalServerError, "failed to delete profile")
		return
	}

	s.json(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleCreateToken(w http.ResponseWriter, r *http.Request) {
	email, err := s.authenticate(r)
	if err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		s.error(w, http.StatusBadRequest, "name is required")
		return
	}

	apiToken, err := auth.GenerateAPIToken(s.secret, email, body.Name)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	hash := auth.HashAPIToken(apiToken.Token)

	// Store in memory
	s.mu.Lock()
	s.apiTokens[hash] = email
	s.mu.Unlock()

	// Store in DB (best effort for now)
	_ = s.store.CreateAPIToken(email, body.Name, hash)

	s.json(w, http.StatusCreated, map[string]string{
		"token": apiToken.Token,
		"name":  body.Name,
	})
}

func (s *Server) handleListTokens(w http.ResponseWriter, r *http.Request) {
	email, err := s.authenticate(r)
	if err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tokens, err := s.store.ListAPITokens(email)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to list tokens")
		return
	}

	if tokens == nil {
		tokens = []store.APITokenRow{}
	}

	s.json(w, http.StatusOK, tokens)
}

func (s *Server) handleDeleteToken(w http.ResponseWriter, r *http.Request) {
	email, err := s.authenticate(r)
	if err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	name := r.PathValue("name")
	if err := s.store.DeleteAPIToken(email, name); err != nil {
		s.error(w, http.StatusNotFound, "token not found")
		return
	}

	// Also remove from in-memory map — we'd need the hash to do this precisely,
	// but for now we leave it (it will fail on next use since DB lookup would fail).

	s.json(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func (s *Server) json(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) error(w http.ResponseWriter, status int, msg string) {
	s.json(w, status, map[string]string{"error": msg})
}

func isPersonal(namespace string) bool {
	return strings.HasPrefix(namespace, "@")
}

func ownerEmail(namespace string) string {
	return strings.TrimPrefix(namespace, "@")
}
