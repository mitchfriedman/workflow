package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type m map[string]interface{}

// httpError is an error that will be rendered to the client.
type httpError struct {
	Message string
	Status  int
}

func Error(status int, message string) *httpError {
	return &httpError{
		Status:  status,
		Message: message,
	}
}

// Error implements the error interface.
func (err *httpError) Error() string {
	return fmt.Sprintf("resterror: %d: %s", err.Status, err.Message)
}

func respond(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("content-type", "application/json; charset=utf-8")

	if httpErr, ok := v.(*httpError); ok {
		respondErr(w, httpErr)
		return
	}

	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
}

func respondErr(w http.ResponseWriter, err error) {
	var wp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if httpErr, ok := err.(*httpError); ok {
		w.WriteHeader(httpErr.Status)
		wp.Error.Message = httpErr.Message
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		wp.Error.Message = err.Error()
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(wp); err != nil {
		panic(err)
	}
}
