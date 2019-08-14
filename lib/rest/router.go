package rest

import (
	"net/http"

	"github.com/mitchfriedman/workflow/lib/logging"

	"github.com/mitchfriedman/workflow/lib/run"
	"gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

// Parser parses an http request to see if it should craete a trigger
// for a run. It can perform logic to determine if the trigger needs
// to be created or not based on the incoming request.
type Parser interface {
	Parse(*http.Request) (*run.Trigger, error)
}

// NewRouter creates and returns a configured mux with registered routes.
func NewRouter(serviceName string, s *run.JobStore, rr run.Repo, p []Parser, logger logging.StructuredLogger) *mux.Router {
	router := mux.NewRouter(mux.WithServiceName(serviceName))
	router.HandleFunc("/healthcheck", BuildHealthcheckHandler()).Methods("GET")
	router.HandleFunc("/Jobs", BuildGetJobsHandler(s)).Methods("GET")
	router.HandleFunc("/Runs", BuildGetRunsHandler(rr)).Methods("GET")
	router.HandleFunc("/Runs/{uuid}", BuildGetRunHandler(rr, logger)).Methods("GET")
	router.HandleFunc("/Triggers", BuildTriggersHandler(s, rr, p, logger)).Methods("POST")

	return router
}
