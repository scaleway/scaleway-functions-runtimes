package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/scaleway/functions-runtime/authentication"
	"github.com/scaleway/functions-runtime/events"
	"github.com/scaleway/functions-runtime/handler"
)

const (
	defaultPort         = 8080
	defaultUpstreamHost = "http://127.0.0.1"
	defaultUpstreamPort = 8081
	headerTriggerType   = "SCW_TRIGGER_TYPE"
)

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

	requestHandler, err := buildRequestHandler()
	if err != nil {
		// TODO: FORMAT ERROR
		return err
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
		Handler:        http.HandlerFunc(requestHandler),
		// NOTE: we should either set timeouts or make explicit we don't need them
		// see https://ieftimov.com/post/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/
	}
	log.Fatal(s.ListenAndServe())

	return nil
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
		if err := authentication.Authenticate(response, request); err != nil {
			log.Print(err)
			return
		}

		// 2: Check event publisher
		triggerType, err := events.GetTriggerType(request.Header.Get(headerTriggerType))
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		// 3: Format event and context
		event, err := events.FormatEvent(request, triggerType)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		context := events.GetExecutionContext()

		// 4: Execute Handler Based on runtime
		handlerResponse, err := fnInvoker.Execute(event, context)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		defer handlerResponse.Close()

		// Do not try to format HTTP response if trigger is NOT of type HTTP (would be pointless as nobody is waiting for the response)
		if triggerType != events.TriggerTypeHTTP {
			io.WriteString(response, "executed properly") // for a trigger 201 Created might be better, so we default to 200
			return
		}

		// 5: Get statusCode, response body, and headers
		handlerRes, err := handler.GetResponse(handlerResponse)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}

		// 6: Send HTTP response with Handler
		// Set Headers
		for key, value := range handlerRes.Headers {
			response.Header().Set(key, value)
		}

		responseBody := handlerRes.Body
		// If user's handler specifies the parameter isBase64Encoded, we need to transform base64 response to byte array
		if handlerRes.IsBase64Encoded {
			var s string
			if err := json.Unmarshal(responseBody, &s); err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
				return
			}

			base64Binary, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				http.Error(response, err.Error(), http.StatusInternalServerError)
				return
			}

			responseBody = base64Binary
		}

		response.WriteHeader(*handlerRes.StatusCode)
		passHandlerResponse(response, responseBody)
	}, nil
}

func passHandlerResponse(w http.ResponseWriter, body json.RawMessage) {
	if len(body) == 0 {
		return
	}
	// when lambda returns a string as body it expects to return it without json encoding
	if body[0] == '"' {
		var s string
		json.Unmarshal(body, &s)
		io.WriteString(w, s)
	} else {
		w.Write(body)
	}
}
