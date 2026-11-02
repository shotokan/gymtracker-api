package primary

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

type errResponse struct {
	Message string `json:"message"`
}

// extractUserID lee el principalId definido por el JWT Authorizer.
func extractUserID(req events.APIGatewayProxyRequest) string {
	if v, ok := req.RequestContext.Authorizer["principalId"]; ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func jsonResp(code int, body any) (events.APIGatewayProxyResponse, error) {
	if body == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: code,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		}, nil
	}

	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(b),
	}, nil
}
