package primary

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

type errResponse struct {
	Error string `json:"error"`
}

func jsonResp(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	data, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(data),
	}, nil
}
