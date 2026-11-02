// activate.go
package input

import "context"

// ActivateCommand representa la solicitud de activación de un usuario, conteniendo el ID del administrador que la realiza y el del usuario a activar.
type ActivateCommand struct {
	ActorID string
	UserID  string
}

// ActivateUseCase es el puerto primario del dominio user para la activación de usuarios.
type ActivateUseCase interface {
	Execute(ctx context.Context, cmd ActivateCommand) error
}
