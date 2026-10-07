package bcrypt

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher implementa hashing con bcrypt.
// Satisface output.PasswordHasher de cualquier dominio via duck typing.
type PasswordHasher struct{ cost int }

// NewPasswordHasher construye un nuevo PasswordHasher con el costo especificado. El costo determina la complejidad del hashing, siendo 10 un valor comúnmente recomendado.
func NewPasswordHasher(cost int) *PasswordHasher {
	return &PasswordHasher{cost: cost}
}

// Hash genera un hash bcrypt a partir de la contraseña en texto plano. Retorna el hash o un error si ocurre algún problema durante el hashing.
func (h *PasswordHasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}

	return string(b), nil
}

// Compare verifica si la contraseña en texto plano coincide con el hash bcrypt. Retorna nil si coinciden o un error si no coinciden o si ocurre algún problema durante la comparación.
func (h *PasswordHasher) Compare(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return fmt.Errorf("comparing password: %w", err)
	}

	return nil
}
