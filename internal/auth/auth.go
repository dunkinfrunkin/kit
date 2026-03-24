package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Sign(secret, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"email": email,
		"iat":   jwt.NewNumericDate(now),
		"exp":   jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Verify(secret, tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return "", errors.New("missing email claim")
	}

	return email, nil
}
