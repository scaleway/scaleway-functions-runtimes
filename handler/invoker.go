package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/scaleway/functions-runtime/events"
)

const (
	retryInterval = time.Millisecond * 50
)

// CoreRuntimeRequest - Structure for a request from core runtime to sub-runtime with event, context, and handler informations to dynamically import
type CoreRuntimeRequest struct {
	Event       interface{}             `json:"event"`
	Context     events.ExecutionContext `json:"context"`
	HandlerName string                  `json:"handlerName"`
	HandlerPath string                  `json:"handlerPath"`
}

// FunctionInvoker - In charge of running sub-runtime processes, and invoke it with all the necessary informations
// to bootstrap the language-specific wrapper to run function handlers
type FunctionInvoker struct {
	RuntimeBridge   string
	RuntimeBinary   string
	HandlerFilePath string
	HandlerName     string
	IsBinary        bool
	client          *http.Client
	upstreamURL     string
}

// NewInvoker - Initialize runtime configuration to execute function handler
// runtimeBinaryPath - Absolute Path to runtime binary (e.g. /usr/bin/python3, /usr/bin/node)
// runtimeBridgePath - Absolute Path to runtime bridge script to start sub-runtime (e.g. /home/app/index.js)
// handlerFilePath - Absolute Path to function handler file (e.g. /home/app/function/myFunction.js for JavaScript or /home/app/function/myHandler for a binary file)
// handlerName - Name of the exported function to use as a Handler (Only for non-compiled languages) to dynamically import function (e.g. handler)
// upstreamURL - URL to sub-runtime HTTP server (e.g. http://localhost:8081)
// isBinaryHandler - Wether function Handler is a binary (Compiled languages)
func NewInvoker(runtimeBinaryPath, runtimeBridgePath, handlerFilePath, handlerName, upstreamURL string, isBinaryHandler bool) (fn *FunctionInvoker, err error) {
	// Need binary path => /usr/local/bin/python3
	// Need runtime bridgle file path => /home/app/runtimes/python3/index.py
	runtimeBinary := runtimeBinaryPath
	runtimeBridgeFile := runtimeBridgePath
	handlerIsBinary := isBinaryHandler

	return &FunctionInvoker{
		RuntimeBridge:   runtimeBridgeFile,
		RuntimeBinary:   runtimeBinary,
		HandlerFilePath: handlerFilePath,
		HandlerName:     handlerName,
		IsBinary:        handlerIsBinary,
		client:          &http.Client{},
		upstreamURL:     upstreamURL,
	}, nil
}

// Start - a new process starting server
func (fn *FunctionInvoker) Start() error {
	var cmd *exec.Cmd
	// If Handler is a binary file, execute binary instead of bridge, and only pass event/context instead of full handler file/name
	if fn.IsBinary {
		cmd = exec.Command(fn.HandlerFilePath)
	} else {
		cmd = exec.Command(fn.RuntimeBinary, fn.RuntimeBridge)
	}

	var (
		stdoutPipe io.ReadCloser
		stdinErr   error
		stdoutErr  error
	)

	_, stdinErr = cmd.StdinPipe()
	if stdinErr != nil {
		return stdinErr
	}

	stdoutPipe, stdoutErr = cmd.StdoutPipe()
	if stdoutErr != nil {
		return stdoutErr
	}

	errPipe, _ := cmd.StderrPipe()

	// Logs lines from stderr and stdout to the stderr and stdout of this process
	bindLoggingPipe("stderr", errPipe, os.Stderr)
	bindLoggingPipe("stdout", stdoutPipe, os.Stdout)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM)

		<-sig
		cmd.Process.Signal(syscall.SIGTERM)

	}()

	err := cmd.Start()
	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Fatalf("Forked function has terminated: %s", err.Error())
		}
	}()

	return err
}

// Execute - a given function handler, and handle response
func (fn *FunctionInvoker) Execute(event interface{}, context events.ExecutionContext) (io.ReadCloser, error) {
	reqBody := CoreRuntimeRequest{
		Event:       event,
		Context:     context,
		HandlerName: fn.HandlerName,
		HandlerPath: fn.HandlerFilePath,
	}

	res, err := fn.streamRequest(reqBody)
	if err != nil {
		return nil, err
	}

	// If an error occured in sub-runtime
	if res.StatusCode == http.StatusInternalServerError {
		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("Read response body error, %v", err)
			return nil, err
		}
		// Error message is the response body
		return nil, handlerExecutionError(string(responseBody))
	}

	return res.Body, nil
}

func (fn FunctionInvoker) streamRequest(reqBody CoreRuntimeRequest) (res *http.Response, err error) {
	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(bodyJSON)
	request, _ := http.NewRequest("POST", fn.upstreamURL, body)
	request.Header.Set("Content-Type", "application/json")

	// Try again, if cold-start, sub-runtime may still be starting-up, try for next 10 seconds, or until it responds properly
	done := false
	retries := 0
	for !done && retries < 200 {
		res, err = fn.client.Do(request)
		if err != nil {
			time.Sleep(retryInterval)
			retries++
			continue
		}
		done = true
	}

	// An error occured
	if !done {
		return nil, fmt.Errorf("too many retries, sub-runtime server did not come up in %v seconds", retryInterval/1000*200)
	}
	return
}
