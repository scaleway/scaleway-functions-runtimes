package server

import (
	"errors"
	"fmt"
	"github.com/scaleway/functions-runtime/authentication"
	"github.com/scaleway/functions-runtime/events"
	"github.com/scaleway/functions-runtime/handler"
	"github.com/scaleway/functions-runtime/utils"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	defaultPort = 8080
	defaultUpstreamHost = "http://127.0.0.1"
	defaultUpstreamPort = 8081
	headerTriggerType = "SCW_TRIGGER_TYPE"
)

var errorInvalidResponseBody = errors.New("Unable to transform HTTP response body")


// Configure function Invoker from environment variables
func setUpFunctionInvoker() (*handler.FunctionInvoker, error) {
	// Exported function to execute
	handlerName := os.Getenv("SCW_HANDLER_NAME")
	// Absolute path to handler file
	handlerPath := os.Getenv("SCW_HANDLER_PATH")
	// Absolute path to runtime binary (e.g. python, node)
	runtimeBinary := os.Getenv("SCW_RUNTIME_BINARY")
	// Absolute path to sub-runtime file
	runtimeBridgeFile := os.Getenv("SCW_RUNTIME_BRIDGE")
	// Whether handler is binary or not (mostly used for compiled languages)
	isBinaryHandler := os.Getenv("SCW_HANDLER_IS_BINARY")
	// Host/Port for sub-runtime HTTP server
	upstreamPort := os.Getenv("SCW_UPSTREAM_PORT")
	upstreamHost := os.Getenv("SCW_UPSTREAM_HOST")

	// Configure connection to upstream server (Function runtime's server)
	if upstreamHost == "" {
		upstreamHost = defaultUpstreamHost
	}
	if upstreamPort == "" {
		upstreamPort = strconv.Itoa(defaultUpstreamPort)
	}
	upstreamURL := fmt.Sprintf("%s:%s", upstreamHost, upstreamPort)

	fnInvoker, err := handler.NewInvoker(runtimeBinary, runtimeBridgeFile, handlerPath, handlerName, upstreamURL, isBinaryHandler == "true")
	if err != nil {
		return nil, err
	}

	return fnInvoker, nil
}

// Start takes the function Handler, at the moment only supporting HTTP Triggers (Api Gateway Proxy events)
// It takes care of wrapping the handler with an HTTP server, which receives requests when functions are triggered
// And execute the handler after formatting the HTTP CoreRuntimeRequest to an API Gateway Proxy Event
func Start() error {
	portEnv := os.Getenv("PORT")
	port, err := strconv.Atoi(portEnv)
	if err != nil {
		port = defaultPort
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	requestHandler, err := buildRequestHandler()
	if err != nil {
		// TODO: FORMAT ERROR
		return err
	}

	http.HandleFunc("/", requestHandler)
	log.Fatal(s.ListenAndServe())

	return nil
}

// Helper function to write HTTP response with given Status Code and Response body
func writeResponse(response http.ResponseWriter, statusCode int, body string) {
	response.WriteHeader(statusCode)
	response.Write([]byte(body))
}

func buildRequestHandler() (func(http.ResponseWriter, *http.Request), error) {
	fnInvoker, err := setUpFunctionInvoker()
	if err != nil {
		return nil, err
	}

	// Start function server
	if err := fnInvoker.Start(); err != nil {
		return nil, err
	}

	return func(response http.ResponseWriter, request *http.Request) {
		// Allow CORS
		response.Header().Set("Access-Control-Allow-Origin", "*")
		response.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Access log
		log.Print("Function Triggered")
		// 1: Authenticate
		// Authenticate function, if an error occurs, do not execute the handler
		if err := authentication.Authenticate(request); err != nil {
			log.Print(err)
			writeResponse(response, http.StatusNotFound, "")
			return
		}

		// 2: Check event publisher
		triggerType, err := events.GetTriggerType(request.Header.Get(headerTriggerType))
		if err != nil {
			writeResponse(response, http.StatusBadRequest, err.Error())
			return
		}

		// 3: Format event and context
		event, err := events.FormatEvent(request, triggerType)
		if err != nil {
			writeResponse(response, http.StatusInternalServerError, err.Error())
			return
		}
		context := events.GetExecutionContext()

		// 4: Execute Handler Based on runtime
		handlerResponse, err := fnInvoker.Execute(event, context)
		if err != nil {
			writeResponse(response, http.StatusInternalServerError, err.Error())
			return
		}

		// Do not try to format HTTP response if trigger is NOT of type HTTP (would be pointless as nobody is waiting for the response)
		if triggerType != events.TriggerTypeHTTP {
			writeResponse(response, http.StatusOK, "executed properly")
			return
		}

		// 5: Get statusCode, response body, and headers
		handlerRes, err := handler.GetResponse(handlerResponse)
		if err != nil {
			writeResponse(response, http.StatusInternalServerError, err.Error())
			return
		}

		// 6: Send HTTP response with Handler
		// Set Headers
		for key, value := range handlerRes.Headers {
			response.Header().Set(key, value)
		}

		// Manage the case where user returns either a string (or stringified JSON) response body, or any other type
		responseBody, err := utils.GetStringFromInterface(handlerRes.Body)
		if err != nil {
			writeResponse(response, http.StatusInternalServerError, errorInvalidResponseBody.Error())
			return
		}

		writeResponse(response, handlerRes.StatusCode, responseBody)
	}, nil
}
