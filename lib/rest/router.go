package rest

import (
	"github.com/mitchfriedman/workflow/lib/run"
	"gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
)

// NewRouter creates and returns a configured mux with registered routes.
func NewRouter(serviceName string, s *run.JobStore, rr run.Repo, p parser) *mux.Router {
	router := mux.NewRouter(mux.WithServiceName(serviceName))
	router.HandleFunc("/Health", BuildHealthcheckHandler())
	router.HandleFunc("/Jobs", BuildGetJobsHandler(s))
	router.HandleFunc("/Runs", BuildGetRunsHandler(rr))
	router.HandleFunc("/Triggers", BuildTriggersHandler(s, rr, p))

	return router
}

