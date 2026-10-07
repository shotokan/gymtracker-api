package domain

// TokenClaims es el value object que representa la identidad
// extraída de un token válido.
type TokenClaims struct {
	UserID string
	Email  string
}
