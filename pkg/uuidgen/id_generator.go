package uuidgen

import "github.com/google/uuid"

// IDGenerator implementa generación de UUIDs v4.
// Satisface output.IDGenerator de cualquier dominio via duck typing.
type IDGenerator struct{}

// New construye un nuevo IDGenerator. No requiere configuración, por lo que no se le pasan parámetros.
func New() *IDGenerator { return &IDGenerator{} }

// New genera un nuevo UUID v4 y lo retorna como string. Retorna un error si ocurre algún problema durante la generación del UUID.
func (g *IDGenerator) New() string { return uuid.New().String() }
