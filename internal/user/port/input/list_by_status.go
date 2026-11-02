// list_by_status.go
package input

import "context"

// ListByStatusCommand representa la solicitud de listado de usuarios, conteniendo el ID del administrador que la realiza y el estado a consultar.
type ListByStatusCommand struct {
	ActorID string
	Status  string
}

// UserSummary representa los datos de un usuario que un administrador necesita para gestionar su acceso.
type UserSummary struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListByStatusResult representa el resultado del listado, ordenado del usuario más antiguo al más reciente.
type ListByStatusResult struct {
	Users []UserSummary `json:"users"`
}

// ListByStatusUseCase es el puerto primario del dominio user para listar usuarios por estado.
type ListByStatusUseCase interface {
	Execute(ctx context.Context, cmd ListByStatusCommand) (ListByStatusResult, error)
}
