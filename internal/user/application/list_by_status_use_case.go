// list_by_status_use_case.go
package application

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/shotokan/gymtracker-api/internal/user/port/input"
	"github.com/shotokan/gymtracker-api/internal/user/port/output"
)

type listByStatusUseCase struct{ repo output.UserRepository }

// NewListByStatus construye el caso de uso de listado de usuarios por estado inyectando su dependencia.
func NewListByStatus(repo output.UserRepository) input.ListByStatusUseCase {
	return &listByStatusUseCase{repo: repo}
}

// Execute retorna los usuarios con el estado indicado si quien lo solicita es un administrador activo.
func (uc *listByStatusUseCase) Execute(ctx context.Context, cmd input.ListByStatusCommand) (input.ListByStatusResult, error) {
	if err := requireActiveAdmin(ctx, uc.repo, cmd.ActorID); err != nil {
		return input.ListByStatusResult{}, err
	}

	users, err := uc.repo.FindByStatus(ctx, cmd.Status)
	if err != nil {
		return input.ListByStatusResult{}, fmt.Errorf("finding users by status: %w", err)
	}

	sort.SliceStable(users, func(i, j int) bool { return users[i].CreatedAt.Before(users[j].CreatedAt) })

	result := input.ListByStatusResult{Users: make([]input.UserSummary, 0, len(users))}
	for _, u := range users {
		result.Users = append(result.Users, input.UserSummary{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Status:    u.Status,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}
