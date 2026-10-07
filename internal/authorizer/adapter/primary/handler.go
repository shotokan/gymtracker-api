package primary

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/shotokan/gymtracker-api/internal/authorizer/domain"
	"github.com/shotokan/gymtracker-api/internal/authorizer/port/input"
)

const bearerPrefix = "Bearer "

// Handler adapta API Gateway Authorizer al caso de uso.
type Handler struct {
	useCase input.AuthorizeUseCase
}

// NewHandler construye el handler del authorizer.
func NewHandler(uc input.AuthorizeUseCase) *Handler {
	return &Handler{useCase: uc}
}

// Handle procesa la solicitud de autorización, extrayendo el token y delegando al caso de uso.
func (h *Handler) Handle(
	ctx context.Context,
	req events.APIGatewayCustomAuthorizerRequest,
) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := extractToken(req.AuthorizationToken)
	if token == "" {
		return deny(req.MethodArn), nil
	}

	result, err := h.useCase.Execute(ctx, input.AuthorizeCommand{Token: token})
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return deny(req.MethodArn), nil
		}
		return deny(req.MethodArn), nil
	}

	return allow(result.UserID, req.MethodArn), nil
}

func extractToken(header string) string {
	if !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}
	return strings.TrimPrefix(header, bearerPrefix)
}

func wildcardArn(methodArn string) string {
	parts := strings.Split(methodArn, ":")
	apiParts := strings.Split(parts[5], "/")
	parts[5] = apiParts[0] + "/" + apiParts[1] + "/*/*"
	return strings.Join(parts, ":")
}

func allow(principalID, arn string) events.APIGatewayCustomAuthorizerResponse {
	return policy(principalID, "Allow", wildcardArn(arn))
}

func deny(arn string) events.APIGatewayCustomAuthorizerResponse {
	return policy("unauthorized", "Deny", arn)
}

func policy(principalID, effect, arn string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{Action: []string{"execute-api:Invoke"}, Effect: effect, Resource: []string{arn}},
			},
		},
	}
}
