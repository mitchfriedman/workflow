package rest

import (
	"github.com/mitchfriedman/workflow/lib/run"
	"net/http"
)

func BuildGetJobsHandler(s *run.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs := s.Jobs()

		resp := m{"jobs": jobs}
		respond(w, http.StatusOK, resp)
	}
}
