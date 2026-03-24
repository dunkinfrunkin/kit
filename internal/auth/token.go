package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type APIToken struct {
	Token     string    `json:"token"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func GenerateAPIToken(secret, email, name string) (*APIToken, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}

	token := "kit_" + hex.EncodeToString(b)

	return &APIToken{
		Token:     token,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

func HashAPIToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
