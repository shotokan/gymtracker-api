package input

import "context"

// AuthorizeCommand representa la solicitud de autorización, conteniendo el token a validar.
type AuthorizeCommand struct {
	Token string
}

// AuthorizeResult representa el resultado de la autorización, indicando el ID del usuario y si se le permite el acceso.
type AuthorizeResult struct {
	UserID  string
	Allowed bool
}

// AuthorizeUseCase es el puerto primario del dominio authorizer.
type AuthorizeUseCase interface {
	Execute(ctx context.Context, cmd AuthorizeCommand) (AuthorizeResult, error)
}
