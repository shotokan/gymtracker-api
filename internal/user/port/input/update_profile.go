// update_profile.go
package input

import "context"

// UpdateProfileCommand representa la solicitud de actualización del perfil de usuario, conteniendo el ID de quien la realiza, el del usuario y el nuevo nombre.
type UpdateProfileCommand struct {
	ActorID string
	UserID  string
	Name    string
}

// UpdateProfileResult representa el resultado de la actualización del perfil de usuario, incluyendo el ID, email y nombre actualizado.
type UpdateProfileResult struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UpdateProfileUseCase es el puerto primario del dominio user para actualizar el perfil de usuario.
type UpdateProfileUseCase interface {
	Execute(ctx context.Context, cmd UpdateProfileCommand) (UpdateProfileResult, error)
}
