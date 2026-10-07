package primary

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type errResponse struct {
	Message string `json:"message"`
}

// extractUserID reads the principalId set by the JWT Authorizer.
func extractUserID(req events.APIGatewayProxyRequest) string {
	if v, ok := req.RequestContext.Authorizer["principalId"]; ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func jsonResp(code int, body any) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Access-Control-Allow-Origin": "*"}

	if body == nil {
		return events.APIGatewayProxyResponse{StatusCode: code, Headers: headers}, nil
	}

	b, _ := json.Marshal(body)
	headers["Content-Type"] = "application/json"

	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Headers:    headers,
		Body:       string(b),
	}, nil
}

var _ = http.StatusOK // keep net/http import if needed elsewhere
