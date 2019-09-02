package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/mitchfriedman/workflow/lib/tracing"

	"github.com/mitchfriedman/workflow/lib/logging"

	"github.com/gorilla/mux"

	"github.com/mitchfriedman/workflow/lib/run"
)

// RunRepresentation is a JSON API response of a run
type RunRepresentation struct {
	ClaimedBy    *string       `json:"claimed_by"`
	ClaimedUntil *time.Time    `json:"claimed_until"`
	CurrentStep  string        `json:"current_step"`
	Finished     *time.Time    `json:"finished"`
	Input        run.InputData `json:"input"`
	Job          string        `json:"job"`
	Rollback     bool          `json:"rollback"`
	Scope        string        `json:"scope"`
	Started      time.Time     `json:"started"`
	State        string        `json:"state"`
	Steps        *run.Step     `json:"steps"`
	UUID         string        `json:"uuid"`
}

func createRepresentation(runs []*run.Run) ([]RunRepresentation, error) {
	reps := make([]RunRepresentation, len(runs), len(runs))
	for i, r := range runs {
		r := r
		rep, err := createRunRepresentation(r)
		if err != nil {
			return nil, err
		}
		reps[i] = rep
	}

	return reps, nil
}

func createRunRepresentation(r *run.Run) (RunRepresentation, error) {
	current, _, err := r.NextStep()
	if err != nil {
		return RunRepresentation{}, errors.Wrap(err, "failed to calculate run NextStep")
	}

	var currentStep string
	if current != nil {
		currentStep = fmt.Sprintf("%s_%s", current.StepType, current.UUID)
	}

	if err := r.UnmarshalRunData(); err != nil {
		return RunRepresentation{}, errors.Wrap(err, "failed to unmarshal run data")
	}

	return RunRepresentation{
		ClaimedBy:    r.ClaimedBy,
		ClaimedUntil: r.ClaimedUntil,
		CurrentStep:  currentStep,
		Finished:     r.Finished,
		Input:        r.Input,
		Job:          r.JobName,
		Rollback:     r.Rollback,
		Scope:        r.Scope,
		Started:      r.Started,
		State:        string(r.State),
		UUID:         r.UUID,
		Steps:        r.Steps,
	}, nil
}

// BuildGetRunHandler builds a HandlerFunc to get a run by the runs UUID.
func BuildGetRunHandler(rr run.Repo, logger logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		uuid := params["uuid"]
		span, ctx := tracing.NewServiceSpan(r.Context(), "get_run")
		defer span.Finish()
		span.SetTag("uuid", uuid)

		found, err := rr.GetRun(ctx, uuid)
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
		res, err := createRunRepresentation(found)
		if err != nil {
			span.RecordError(err)
			logger.Errorf("failed to marshal run %v", found, err)
			respondErr(w, Error(http.StatusInternalServerError, err.Error()))
			return
		}

		respond(w, http.StatusOK, res)
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

		res, err := createRepresentation(runs)
		if err != nil {
			span.RecordError(err)
			respondErr(w, Error(http.StatusInternalServerError, err.Error()))
			return
		}

		result := m{"runs": res}
		respond(w, http.StatusOK, result)
	}
}
