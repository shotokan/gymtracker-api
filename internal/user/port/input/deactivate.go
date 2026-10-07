// deactivate.go
package input

import "context"

// DeactivateCommand representa la solicitud de desactivación de un usuario, conteniendo el ID del usuario a desactivar.
type DeactivateCommand struct {
	UserID string
}

// DeactivateUseCase es el puerto primario del dominio user para la desactivación de usuarios.
type DeactivateUseCase interface {
	Execute(ctx context.Context, cmd DeactivateCommand) error
}
