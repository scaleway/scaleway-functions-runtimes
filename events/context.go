package events

import (
	"os"
	"strconv"
)

var (
	memoryLimitInMb int
	functionName string
	functionVersion string
)

const (
	defaultMemoryInMB = 128
)

// ExecutionContext - type for the context of execution of the function including memory, function name and version...
type ExecutionContext struct{
	MemoryLimitInMB int `json:"memoryLimitInMb"`
	FunctionName string `json:"functionName"`
	FunctionVersion string `json:"functionVersion"`
}

// GetExecutionContext - retrieve the execution context of the current function
func GetExecutionContext() ExecutionContext {
	return ExecutionContext{
		MemoryLimitInMB: memoryLimitInMb,
		FunctionName: functionName,
		FunctionVersion: functionVersion,
	}
}

func init() {
	var err error
	memoryLimitInMb, err = strconv.Atoi(os.Getenv("SCW_APPLICATION_MEMORY"))
	if err != nil {
		memoryLimitInMb = defaultMemoryInMB
	}

	functionName = os.Getenv("SCW_APPLICATION_NAME")
	functionVersion = os.Getenv("SCW_APPLICATION_VERSION")
}
