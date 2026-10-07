package application

import (
	"context"
	"errors"
	"testing"

	"github.com/shotokan/gymtracker-api/internal/plan/domain"
	"github.com/shotokan/gymtracker-api/internal/plan/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mocks ─────────────────────────────────────────────────────────────────

type mockPlanRepo struct{ mock.Mock }

func (m *mockPlanRepo) Save(ctx context.Context, plan *domain.Plan) error {
	return m.Called(ctx, plan).Error(0)
}

func (m *mockPlanRepo) FindByID(ctx context.Context, userID, planID string) (*domain.Plan, error) {
	args := m.Called(ctx, userID, planID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Plan), args.Error(1)
}

func (m *mockPlanRepo) FindAll(ctx context.Context, userID, nameFilter string) ([]domain.Plan, error) {
	args := m.Called(ctx, userID, nameFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Plan), args.Error(1)
}

func (m *mockPlanRepo) FindActive(ctx context.Context, userID string) (*domain.Plan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Plan), args.Error(1)
}

func (m *mockPlanRepo) Update(ctx context.Context, plan *domain.Plan) error {
	return m.Called(ctx, plan).Error(0)
}

func (m *mockPlanRepo) Delete(ctx context.Context, userID, planID string) error {
	return m.Called(ctx, userID, planID).Error(0)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) New() string {
	return m.Called().String(0)
}

// ── Tests ─────────────────────────────────────────────────────────────────

func TestCreatePlanUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	idGen := new(mockIDGen)

	idGen.On("New").Return("plan-123")
	repo.On("Save", mock.Anything, mock.MatchedBy(func(p *domain.Plan) bool {
		return p.ID == "plan-123" &&
			p.UserID == "user-1" &&
			p.Name == "Mi Plan" &&
			!p.Active &&
			len(p.Routines) == 0
	})).Return(nil)

	uc := NewCreatePlan(repo, idGen)

	result, err := uc.Execute(context.Background(), input.CreatePlanCommand{
		UserID: "user-1",
		Name:   "Mi Plan",
	})

	require.NoError(t, err)
	assert.Equal(t, "plan-123", result.ID)
	assert.Equal(t, "Mi Plan", result.Name)
	assert.False(t, result.Active)
	assert.Empty(t, result.Routines)
	repo.AssertExpectations(t)
	idGen.AssertExpectations(t)
}

func TestCreatePlanUseCase_Execute_RepositoryError(t *testing.T) {
	t.Parallel()

	repo := new(mockPlanRepo)
	idGen := new(mockIDGen)

	idGen.On("New").Return("plan-123")
	repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("dynamo error"))

	uc := NewCreatePlan(repo, idGen)

	_, err := uc.Execute(context.Background(), input.CreatePlanCommand{
		UserID: "user-1",
		Name:   "Mi Plan",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "saving plan")
}
