package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// ResponseHTTP - Type for HTTP triggers response emitted by function handlers
type ResponseHTTP struct {
	StatusCode      int               `json:"statusCode"`
	Body            json.RawMessage   `json:"body"`
	Headers         map[string]string `json:"headers"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

// GetResponse - Transform a response string into an HTTP Response structure
func GetResponse(response io.Reader) (*ResponseHTTP, error) {
	handlerResponse := &ResponseHTTP{}

	// Read body content
	bodyBytes, err := ioutil.ReadAll(response)
	if err != nil {
		return nil, ErrorInvalidHTTPResponseFormat
	}

	err = json.Unmarshal(bodyBytes, &handlerResponse)
	if err != nil {
		// Body is not a valid JSON object
		handlerResponse.Body = bodyBytes
	}

	// If handler dit not return status code, just use 200 OK
	if handlerResponse.StatusCode == 0 {
		handlerResponse.StatusCode = http.StatusOK
	}

	return handlerResponse, nil
}
