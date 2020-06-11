package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/scaleway/functions-runtime/handler"
)

func Test_passHandlerResponse_string(t *testing.T) {
	r := &handler.ResponseHTTP{}
	json.Unmarshal([]byte(`
			{
         "body":  "this is a string   "  
      }
`), r)
	recorder := httptest.NewRecorder()
	passHandlerResponse(recorder, r.Body)
	if recorder.Body.String() != "this is a string   " {
		t.Fail()
	}
}

func Test_passHandlerResponse_number(t *testing.T) {
	r := &handler.ResponseHTTP{}
	json.Unmarshal([]byte(`
			{
         "body": 5
      }
`), r)
	recorder := httptest.NewRecorder()
	passHandlerResponse(recorder, r.Body)
	if recorder.Body.String() != "5" {
		t.Fail()
	}
}

func Test_passHandlerResponse_bool(t *testing.T) {
	r := &handler.ResponseHTTP{}
	json.Unmarshal([]byte(`
			{
         "body": true
      }
`), r)
	recorder := httptest.NewRecorder()
	passHandlerResponse(recorder, r.Body)
	if recorder.Body.String() != "true" {
		t.Fail()
	}
}

func Test_passHandlerResponse_json(t *testing.T) {
	r := &handler.ResponseHTTP{}
	json.Unmarshal([]byte(`
			{
         "body":  {  "a": 2,  "4": "asdds" }
      }
`), r)
	recorder := httptest.NewRecorder()
	passHandlerResponse(recorder, r.Body)
	if recorder.Body.String() != "{  \"a\": 2,  \"4\": \"asdds\" }" {
		t.Fail()
	}
}

func Test_passHandlerResponse_string_json(t *testing.T) {
	r := &handler.ResponseHTTP{}
	json.Unmarshal([]byte(`
			{
         "body": "{  \"a\": 2,  \"4\": \"asdds\" }"
      }
`), r)
	recorder := httptest.NewRecorder()
	passHandlerResponse(recorder, r.Body)
	if recorder.Body.String() != "{  \"a\": 2,  \"4\": \"asdds\" }" {
		t.Fail()
	}
}
