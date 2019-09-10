package events

import (
	"errors"
	"io/ioutil"
	"net/http"
)

var (
	// TriggerTypeMQTT - Event trigger of type MQTT - pub/sub
	TriggerTypeMQTT TriggerType = "mqtt"
	// TriggerTypeHTTP - Event trigger of type HTTP
	TriggerTypeHTTP TriggerType = "http"
	// ValidTriggerTypes - List of supported trigger types
	ValidTriggerTypes = []TriggerType{TriggerTypeMQTT}
	// ErrorNotSupportedTrigger - Error when event is assigned to not supported trigger types
	ErrorNotSupportedTrigger = errors.New("Trigger Type is not supported by Scaleway Functions Runtime")
)

// TriggerType - Enumeration of valid trigger types supported by runtime
type TriggerType string

// GetTriggerType - check that a given trigger type is supported by runtime
func GetTriggerType(triggerType string) (trigger TriggerType, err error) {
	if triggerType == "" {
		return TriggerTypeHTTP, nil
	}

	for _, validType := range ValidTriggerTypes {
		if string(validType) == triggerType {
			return validType, nil
		}
	}

	return "", ErrorNotSupportedTrigger
}

// FormatEvent - Format event according to given trigger type, if trigger type if not HTTP, then we assume that event
// has already been formatted by event-source
func FormatEvent(req *http.Request, triggerType TriggerType) (interface{}, error) {
	if triggerType == TriggerTypeHTTP {
		return formatEventHTTP(req), nil
	}
	// request body is the event
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, errors.New("Unable to read request body")
	}

	return string(reqBody), nil
}
