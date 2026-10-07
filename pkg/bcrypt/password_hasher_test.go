package bcrypt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shotokan/gymtracker-api/pkg/bcrypt"
)

func TestPasswordHasher(t *testing.T) {
	t.Parallel()

	h := bcrypt.NewPasswordHasher(4) // cost bajo para tests rápidos

	t.Run("hash y compare exitoso", func(t *testing.T) {
		t.Parallel()
		hash, err := h.Hash("mysecret")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		require.NoError(t, h.Compare(hash, "mysecret"))
	})

	t.Run("contraseña incorrecta → error", func(t *testing.T) {
		t.Parallel()
		hash, _ := h.Hash("mysecret")
		assert.Error(t, h.Compare(hash, "wrong"))
	})

	t.Run("cada hash es único (salt)", func(t *testing.T) {
		t.Parallel()
		h1, _ := h.Hash("same")
		h2, _ := h.Hash("same")
		assert.NotEqual(t, h1, h2)
	})
}
