package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shotokan/gymtracker-api/internal/authorizer/application"
	"github.com/shotokan/gymtracker-api/internal/authorizer/domain"
	"github.com/shotokan/gymtracker-api/internal/authorizer/port/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockValidator struct{ mock.Mock }

func (m *mockValidator) Validate(ctx context.Context, token string) (*domain.TokenClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenClaims), args.Error(1)
}

func TestAuthorizeUseCase_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		token   string
		setup   func(*mockValidator)
		want    input.AuthorizeResult
		wantErr error
	}{
		{
			name:  "token válido → Allowed",
			token: "valid.jwt",
			setup: func(m *mockValidator) {
				m.On("Validate", mock.Anything, "valid.jwt").
					Return(&domain.TokenClaims{UserID: "u-1", Email: "ivan@gym.com"}, nil)
			},
			want: input.AuthorizeResult{UserID: "u-1", Allowed: true},
		},
		{
			name:  "token inválido → ErrUnauthorized",
			token: "bad.jwt",
			setup: func(m *mockValidator) {
				m.On("Validate", mock.Anything, "bad.jwt").Return(nil, errors.New("expired"))
			},
			wantErr: domain.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := new(mockValidator)
			tt.setup(v)

			uc := application.New(v)
			result, err := uc.Execute(context.Background(), input.AuthorizeCommand{Token: tt.token})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}
