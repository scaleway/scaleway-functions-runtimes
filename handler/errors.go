package handler

import (
	"errors"
	"fmt"
)

var (
	// ErrorInvalidHTTPResponseFormat - Error type for mal-formatted responses from user's handlers
	ErrorInvalidHTTPResponseFormat = errors.New("Handler's results for HTTP response is mal-formatted")
	// ErrorPayloadTooLarge - Error type for payload size is grater that anticipated
	ErrorPayloadTooLarge = errors.New("Payload too large")
)

func handlerExecutionError(err string) error {
	errorMessage := fmt.Sprintf("An error occured during handler execution: %s", err)
	return errors.New(errorMessage)
}
