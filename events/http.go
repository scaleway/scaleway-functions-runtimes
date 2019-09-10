package events

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
)

// APIGatewayProxyRequest contains data coming from the API Gateway proxy
type APIGatewayProxyRequest struct {
	Resource                        string                        `json:"resource"` // The resource path defined in API Gateway
	Path                            string                        `json:"path"`     // The url path for the caller
	HTTPMethod                      string                        `json:"httpMethod"`
	Headers                         map[string]string             `json:"headers"`
	MultiValueHeaders               map[string][]string           `json:"multiValueHeaders"`
	QueryStringParameters           map[string]string             `json:"queryStringParameters"`
	MultiValueQueryStringParameters map[string][]string           `json:"multiValueQueryStringParameters"`
	PathParameters                  map[string]string             `json:"pathParameters"`
	StageVariables                  map[string]string             `json:"stageVariables"`
	RequestContext                  APIGatewayProxyRequestContext `json:"requestContext"`
	Body                            string                        `json:"body"`
	IsBase64Encoded                 bool                          `json:"isBase64Encoded,omitempty"`
}

// APIGatewayProxyRequestContext contains the information to identify the AWS account and resources invoking the
// Lambda function. It also includes Cognito identity information for the caller.
type APIGatewayProxyRequestContext struct {
	AccountID    string                 `json:"accountId"`
	ResourceID   string                 `json:"resourceId"`
	Stage        string                 `json:"stage"`
	RequestID    string                 `json:"requestId"`
	ResourcePath string                 `json:"resourcePath"`
	Authorizer   map[string]interface{} `json:"authorizer"`
	HTTPMethod   string                 `json:"httpMethod"`
	APIID        string                 `json:"apiId"` // The API Gateway rest API Id
}

func formatEventHTTP(r *http.Request) APIGatewayProxyRequest {
	var input string

	if r.Body != nil {
		defer r.Body.Close()

		bodyBytes, bodyErr := ioutil.ReadAll(r.Body)

		if bodyErr != nil {
			log.Printf("Error reading body from request.")
		}

		input = string(bodyBytes)
	}

	headers := map[string]string{}
	for key, value := range r.Header {
		headers[key] = value[len(value)-1]
	}

	queryParameters := map[string]string{}
	for key, value := range r.URL.Query() {
		queryParameters[key] = value[len(value)-1]
	}

	isBase64Encoded := true
	_, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		isBase64Encoded = false
	}

	event := APIGatewayProxyRequest{
		Path:                  r.URL.Path,
		HTTPMethod:            r.Method,
		Headers:               headers,
		QueryStringParameters: queryParameters,
		StageVariables:        map[string]string{},
		Body:                  input,
		IsBase64Encoded:       isBase64Encoded,
		RequestContext: APIGatewayProxyRequestContext{
			Stage:      "",
			HTTPMethod: r.Method,
		},
	}

	return event
}
