package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
}

type OIDCVerifier struct {
	config  OIDCConfig
	jwksURL string
	jwks    *jwksCache
}

type jwksCache struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
}

func (v *OIDCVerifier) Config() OIDCConfig {
	return v.config
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type oidcClaims struct {
	Iss               string      `json:"iss"`
	Sub               string      `json:"sub"`
	Aud               interface{} `json:"aud"`
	Exp               int64       `json:"exp"`
	Email             string      `json:"email"`
	PreferredUsername  string      `json:"preferred_username"`
}

func (c *oidcClaims) hasAudience(clientID string) bool {
	switch v := c.Aud.(type) {
	case string:
		return v == clientID
	case []interface{}:
		for _, a := range v {
			if s, ok := a.(string); ok && s == clientID {
				return true
			}
		}
	}
	return false
}

func NewOIDCVerifier(cfg OIDCConfig) (*OIDCVerifier, error) {
	if cfg.Issuer == "" {
		return nil, errors.New("oidc: issuer is required")
	}

	discoveryURL := strings.TrimRight(cfg.Issuer, "/") + "/.well-known/openid-configuration"
	resp, err := http.Get(discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to fetch discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc: discovery endpoint returned status %d", resp.StatusCode)
	}

	var doc struct {
		JwksURI string `json:"jwks_uri"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("oidc: failed to parse discovery document: %w", err)
	}
	if doc.JwksURI == "" {
		return nil, errors.New("oidc: discovery document missing jwks_uri")
	}

	return &OIDCVerifier{
		config:  cfg,
		jwksURL: doc.JwksURI,
		jwks: &jwksCache{
			keys: make(map[string]*rsa.PublicKey),
		},
	}, nil
}

func (v *OIDCVerifier) VerifyToken(tokenStr string) (string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return "", errors.New("oidc: malformed JWT")
	}

	headerBytes, err := base64URLDecode(parts[0])
	if err != nil {
		return "", fmt.Errorf("oidc: failed to decode JWT header: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", fmt.Errorf("oidc: failed to parse JWT header: %w", err)
	}
	if header.Alg != "RS256" {
		return "", fmt.Errorf("oidc: unsupported algorithm %q", header.Alg)
	}

	key, ok := v.jwks.getKey(header.Kid)
	if !ok || time.Since(v.jwks.fetched) > time.Hour {
		if err := v.refreshJWKS(); err != nil {
			return "", fmt.Errorf("oidc: failed to refresh JWKS: %w", err)
		}
		key, ok = v.jwks.getKey(header.Kid)
		if !ok {
			return "", fmt.Errorf("oidc: unknown kid %q", header.Kid)
		}
	}

	sigBytes, err := base64URLDecode(parts[2])
	if err != nil {
		return "", fmt.Errorf("oidc: failed to decode signature: %w", err)
	}

	signed := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(signed))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], sigBytes); err != nil {
		return "", fmt.Errorf("oidc: invalid signature: %w", err)
	}

	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return "", fmt.Errorf("oidc: failed to decode JWT payload: %w", err)
	}

	var claims oidcClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", fmt.Errorf("oidc: failed to parse JWT claims: %w", err)
	}

	if claims.Iss != v.config.Issuer {
		return "", fmt.Errorf("oidc: issuer mismatch: got %q, want %q", claims.Iss, v.config.Issuer)
	}
	if !claims.hasAudience(v.config.ClientID) {
		return "", errors.New("oidc: token audience does not match client ID")
	}
	if time.Now().Unix() > claims.Exp {
		return "", errors.New("oidc: token expired")
	}

	if claims.Email != "" {
		return claims.Email, nil
	}
	if claims.PreferredUsername != "" {
		return claims.PreferredUsername, nil
	}
	if claims.Sub != "" {
		return claims.Sub, nil
	}
	return "", errors.New("oidc: no email, preferred_username, or sub claim found")
}

func (v *jwksCache) getKey(kid string) (*rsa.PublicKey, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	key, ok := v.keys[kid]
	return key, ok
}

func (v *OIDCVerifier) refreshJWKS() error {
	resp, err := http.Get(v.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwksDoc struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwksDoc); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey)
	for _, k := range jwksDoc.Keys {
		if k.Kty != "RSA" {
			continue
		}

		nBytes, err := base64URLDecode(k.N)
		if err != nil {
			return fmt.Errorf("failed to decode modulus for kid %q: %w", k.Kid, err)
		}
		eBytes, err := base64URLDecode(k.E)
		if err != nil {
			return fmt.Errorf("failed to decode exponent for kid %q: %w", k.Kid, err)
		}

		n := new(big.Int).SetBytes(nBytes)
		e := new(big.Int).SetBytes(eBytes)

		keys[k.Kid] = &rsa.PublicKey{
			N: n,
			E: int(e.Int64()),
		}
	}

	v.jwks.mu.Lock()
	v.jwks.keys = keys
	v.jwks.fetched = time.Now()
	v.jwks.mu.Unlock()

	return nil
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
