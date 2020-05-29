package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

// ResponseHTTP - Type for HTTP triggers response emitted by function handlers
type ResponseHTTP struct {
	StatusCode int               `json:"statusCode"`
	Body       json.RawMessage   `json:"body"`
	Headers    map[string]string `json:"headers"`
}

// GetResponse - Transform a response string into an HTTP Response structure
func GetResponse(response io.Reader) (*ResponseHTTP, error) {
	handlerResponse := &ResponseHTTP{}
	if err := json.NewDecoder(response).Decode(handlerResponse); err != nil {
		return nil, ErrorInvalidHTTPResponseFormat
	}

	// If handler dit not return status code, just use 200 OK
	if handlerResponse.StatusCode == 0 {
		handlerResponse.StatusCode = http.StatusOK
	}

	return handlerResponse, nil
}
