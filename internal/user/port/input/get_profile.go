// get_profile.go
package input

import "context"

// GetProfileCommand representa la solicitud de obtención del perfil de usuario, conteniendo el ID del usuario.
type GetProfileCommand struct {
	UserID string
}

// GetProfileResult representa el resultado de la obtención del perfil de usuario, incluyendo su ID, email, nombre, estado activo y fecha de creación.
type GetProfileResult struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

// GetProfileUseCase es el puerto primario del dominio user para la obtención del perfil de usuario.
type GetProfileUseCase interface {
	Execute(ctx context.Context, cmd GetProfileCommand) (GetProfileResult, error)
}
