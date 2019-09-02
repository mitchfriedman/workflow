package rest

import (
	"net/http"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/tracing"
)

func BuildGetJobsHandler(s *run.JobStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs := s.Jobs()

		span, _ := tracing.NewServiceSpan(r.Context(), "get_jobs")
		defer span.Finish()
		resp := m{"jobs": jobs}
		respond(w, http.StatusOK, resp)
	}
}
