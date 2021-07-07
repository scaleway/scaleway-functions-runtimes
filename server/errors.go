package server

import (
	"fmt"
)

// ErrorPayloadTooLarge - Error type for payload size is grater that anticipated
var ErrorPayloadTooLarge = fmt.Errorf("Request payload too large, max payload size = %d bytes", payloadMaxSize)
