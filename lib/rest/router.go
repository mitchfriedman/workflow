package rest

import (
	"net/http"

	"github.com/mitchfriedman/workflow/lib/run"
	"gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

// Parser parses an http request to see if it should craete a trigger
// for a run. It can perform logic to determine if the trigger needs
// to be created or not based on the incoming request.
type Parser interface {
	Parse(*http.Request) (run.Trigger, error)
}

// NewRouter creates and returns a configured mux with registered routes.
func NewRouter(serviceName string, s *run.JobStore, rr run.Repo, p Parser) *mux.Router {
	router := mux.NewRouter(mux.WithServiceName(serviceName))
	router.HandleFunc("/Health", BuildHealthcheckHandler())
	router.HandleFunc("/Jobs", BuildGetJobsHandler(s))
	router.HandleFunc("/Runs", BuildGetRunsHandler(rr))
	router.HandleFunc("/Triggers", BuildTriggersHandler(s, rr, p))

	return router
}
