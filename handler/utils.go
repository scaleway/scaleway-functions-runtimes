package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	httpStatusOK = http.StatusOK
)

// ResponseHTTP - Type for HTTP triggers response emitted by function handlers
type ResponseHTTP struct {
	StatusCode      *int              `json:"statusCode"`
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

	unmarshalErr := json.Unmarshal(bodyBytes, &handlerResponse)

	// If handler dit not return a JSON or status code, just use 200 OK
	if unmarshalErr != nil || handlerResponse.StatusCode == nil {
		handlerResponse.StatusCode = &httpStatusOK
		handlerResponse.Body = bodyBytes
	}

	return handlerResponse, nil
}
