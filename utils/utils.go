package utils

import (
	"encoding/json"
)

// MarshalJSON - Marshal an interface into a stringified JSON structure
func MarshalJSON(structure interface{}) (string, error) {
	byteArray, err := json.Marshal(structure)
	if err != nil {
		return "", err
	}

	return string(byteArray), nil
}

// GetStringFromInterface - Transform an interface into a string
func GetStringFromInterface(toCompute interface{}) (stringified string, err error) {
	var ok bool
	if stringified, ok = toCompute.(string); !ok {
		stringified, err = MarshalJSON(toCompute)
		if err != nil {
			return
		}
	}

	return
}
