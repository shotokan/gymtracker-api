package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shotokan/gymtracker-api/internal/authorizer/adapter/primary"
	authorizerjwt "github.com/shotokan/gymtracker-api/internal/authorizer/adapter/secondary/jwt"
	"github.com/shotokan/gymtracker-api/internal/authorizer/application"
)

// Los handlers identifican al usuario con el principalId del Authorizer: un token de login
// debe llegar a ellos con el ID del usuario.
func TestLoginToken_AuthorizerResolvesUserID(t *testing.T) {
	t.Parallel()

	token, err := NewTokenGenerator("secret", time.Hour).Generate("user-123", "ivan@gym.com")
	require.NoError(t, err)

	validator := authorizerjwt.NewTokenValidator("secret", time.Hour)
	handler := primary.NewHandler(application.New(validator))

	resp, err := handler.Handle(context.Background(), events.APIGatewayCustomAuthorizerRequest{
		AuthorizationToken: "Bearer " + token,
		MethodArn:          "arn:aws:execute-api:us-east-1:123456789012:abc123/v1/GET/plans",
	})

	require.NoError(t, err)
	assert.Equal(t, "user-123", resp.PrincipalID)
	assert.Equal(t, "Allow", resp.PolicyDocument.Statement[0].Effect)
}
