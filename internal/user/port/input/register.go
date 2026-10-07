// register.go
package input

import "context"

// RegisterCommand representa la solicitud de registro, conteniendo el email, contraseña y nombre del usuario.
type RegisterCommand struct {
	Email    string
	Password string
	Name     string
}

// RegisterResult representa el resultado del registro, incluyendo el ID generado para el nuevo usuario, su email y nombre.
type RegisterResult struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// RegisterUseCase es el puerto primario del dominio user para el registro de nuevos usuarios.
type RegisterUseCase interface {
	Execute(ctx context.Context, cmd RegisterCommand) (RegisterResult, error)
}
