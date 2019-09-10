package handler

import (
	"encoding/json"
	"net/http"
)

// ResponseHTTP - Type for HTTP triggers response emitted by function handlers
type ResponseHTTP struct{
	StatusCode int `json:"statusCode"`
	Body interface{} `json:"body"`
	Headers map[string]string `json:"headers"`
}

// GetResponse - Transform a response string into an HTTP Response structure
func GetResponse(response string) (handlerResponse *ResponseHTTP, err error) {
	if err := json.Unmarshal([]byte(response), &handlerResponse); err != nil {
		return nil, ErrorInvalidHTTPResponseFormat
	}

	// If handler dit not return status code, just use 200 OK
	if handlerResponse.StatusCode == 0 {
		handlerResponse.StatusCode = http.StatusOK
	}

	return
}
