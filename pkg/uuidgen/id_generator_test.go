package uuidgen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shotokan/gymtracker-api/pkg/uuidgen"
)

func TestIDGenerator(t *testing.T) {
	t.Parallel()

	g := uuidgen.New()

	t.Run("genera ID no vacío", func(t *testing.T) {
		t.Parallel()
		assert.NotEmpty(t, g.New())
	})

	t.Run("cada ID es único", func(t *testing.T) {
		t.Parallel()
		assert.NotEqual(t, g.New(), g.New())
	})

	t.Run("formato UUID válido (36 chars)", func(t *testing.T) {
		t.Parallel()
		id := g.New()
		assert.Len(t, id, 36)
		assert.Equal(t, uint8('-'), id[8])
	})
}
