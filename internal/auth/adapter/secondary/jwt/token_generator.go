package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenGenerator generates JWT tokens for authenticated users.
type TokenGenerator struct {
	secret     []byte
	expiration time.Duration
}

// NewTokenGenerator creates a new JWT token generator.
// secret is the signing key (should be stored in environment variable).
// expiration is how long the token is valid (e.g., 24 hours).
func NewTokenGenerator(secret string, expiration time.Duration) *TokenGenerator {
	return &TokenGenerator{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

// Claims represents the JWT payload.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Generate creates a signed JWT token for the user.
func (g *TokenGenerator) Generate(userID, email string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID, // the authorizer reads the user ID from sub
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(g.expiration)),
			Issuer:    "gymtracker-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(g.secret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signedToken, nil
}

// Validate verifies a JWT token and returns the claims.
func (g *TokenGenerator) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
