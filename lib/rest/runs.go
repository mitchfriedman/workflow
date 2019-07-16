package rest

import (
	"context"
	"github.com/mitchfriedman/workflow/lib/run"
	"net/http"
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

// BuildGetRunsHandler builds HandlerFunc to fetch runs from the database.
func BuildGetRunsHandler(rr run.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobParam, ok := r.URL.Query()["job_name"]
		if !ok || len(jobParam) < 1 {
			respondErr(w, Error(http.StatusBadRequest, "missing required query parameter: 'job_name'"))
			return
		}
		job := jobParam[0]

		runs, err := rr.ListByJob(context.TODO(), job)
		if err != nil {
			respondErr(w, Error(http.StatusInternalServerError, err.Error()))
			return
		}

		result := m{"runs": createRepresentation(runs)}

		respond(w, http.StatusOK, result)
	}
}
