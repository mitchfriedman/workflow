package rest

import (
	"net/http"

	"github.com/mitchfriedman/workflow/lib/tracing"

	"github.com/mitchfriedman/workflow/lib/logging"

	"github.com/gorilla/mux"

	"github.com/mitchfriedman/workflow/lib/run"
)

// RunRepresentation is a JSON API response of a run
type RunRepresentation struct {
	Run *run.Run `json:"run"`
}

func createRepresentation(runs []*run.Run) []RunRepresentation {
	reps := make([]RunRepresentation, len(runs), len(runs))
	for i, r := range runs {
		r := r
		reps[i] = RunRepresentation{
			Run: r,
		}
	}

	return reps
}

// BuildGetRunHandler builds a HandlerFunc to get a run by the runs UUID.
func BuildGetRunHandler(rr run.Repo, logger logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		uuid := params["uuid"]
		span, ctx := tracing.NewServiceSpan(r.Context(), "get_run")
		defer span.Finish()
		span.SetTag("uuid", uuid)

		result, err := rr.GetRun(ctx, uuid)
		if err != nil {
			switch err {
			case run.ErrNotFound:
				respondErr(w, Error(http.StatusNotFound, err.Error()))
			default:
				span.RecordError(err)
				logger.Errorf("failed to get run with uuid %s - %v", uuid, err)
				respondErr(w, Error(http.StatusInternalServerError, err.Error()))
			}
			return
		}
		respond(w, http.StatusOK, result)
	}
}

// BuildGetRunsHandler builds a HandlerFunc to fetch runs from the database.
func BuildGetRunsHandler(rr run.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobParam, ok := r.URL.Query()["job_name"]
		if !ok || len(jobParam) < 1 {
			respondErr(w, Error(http.StatusBadRequest, "missing required query parameter: 'job_name'"))
			return
		}
		job := jobParam[0]

		ctx := r.Context()
		span, ctx := tracing.NewServiceSpan(ctx, "list_runs_by_job")
		defer span.Finish()
		span.SetTag("job_type", job)

		runs, err := rr.ListByJob(ctx, job)
		if err != nil {
			span.RecordError(err)
			respondErr(w, Error(http.StatusInternalServerError, err.Error()))
			return
		}

		result := m{"runs": createRepresentation(runs)}

		respond(w, http.StatusOK, result)
	}
}
