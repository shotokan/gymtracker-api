package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/shotokan/gymtracker-api/internal/authorizer/domain"
)

type claims struct {
	gojwt.RegisteredClaims
	Email string `json:"email"`
}

// TokenValidator implementa output.TokenValidator con JWT HS256.
type TokenValidator struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenValidator construye un nuevo TokenValidator con el secreto y TTL especificados.|
func NewTokenValidator(secret string, ttl time.Duration) *TokenValidator {
	return &TokenValidator{secret: []byte(secret), ttl: ttl}
}

// Validate verifica la validez del token JWT y extrae las reclamaciones. Retorna error si el token es inválido o ha expirado.
func (v *TokenValidator) Validate(_ context.Context, tokenStr string) (*domain.TokenClaims, error) {
	var c claims

	token, err := gojwt.ParseWithClaims(tokenStr, &c, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return v.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return &domain.TokenClaims{
		UserID: c.Subject,
		Email:  c.Email,
	}, nil
}
