package application

import (
	"context"

	"github.com/shotokan/gymtracker-api/internal/user/domain"
	"github.com/stretchr/testify/mock"
)

// ── Shared Test Mocks ─────────────────────────────────────────────────────

type mockUserRepo struct{ mock.Mock }

//nolint:wrapcheck // Test mocks don't need error wrapping
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

//nolint:wrapcheck // Test mocks don't need error wrapping
func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

//nolint:wrapcheck // Test mocks don't need error wrapping
func (m *mockUserRepo) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	args := m.Called(ctx, googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

//nolint:wrapcheck // Test mocks don't need error wrapping
func (m *mockUserRepo) Save(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

//nolint:wrapcheck // Test mocks don't need error wrapping
func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

type mockPasswordHasher struct{ mock.Mock }

//nolint:wrapcheck // Test mocks don't need error wrapping
func (m *mockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

//nolint:wrapcheck // Test mocks don't need error wrapping
func (m *mockPasswordHasher) Compare(hash, plain string) error {
	args := m.Called(hash, plain)
	return args.Error(0)
}

type mockIDGen struct{ mock.Mock }

func (m *mockIDGen) New() string {
	args := m.Called()
	return args.String(0)
}
