package rest

import (
	"context"
	"net/http"

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
func BuildGetRunHandler(rr run.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		uuid := params["uuid"]
		result, err := rr.GetRun(context.Background(), uuid)
		if err != nil {
			switch err {
			case run.ErrNotFound:
				respondErr(w, Error(http.StatusNotFound, err.Error()))
			default:
				respondErr(w, Error(http.StatusInternalServerError, err.Error()))
			}
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

		runs, err := rr.ListByJob(context.TODO(), job)
		if err != nil {
			respondErr(w, Error(http.StatusInternalServerError, err.Error()))
			return
		}

		result := m{"runs": createRepresentation(runs)}

		respond(w, http.StatusOK, result)
	}
}
