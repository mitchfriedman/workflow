package rest

import (
	"context"
	"net/http"

	"github.com/mitchfriedman/workflow/lib/tracing"

	"github.com/mitchfriedman/workflow/lib/logging"

	"github.com/mitchfriedman/workflow/lib/run"
)

func parse(parsers []Parser, req *http.Request) (*run.Trigger, error) {
	for _, p := range parsers {
		trig, err := p.Parse(req)
		if err != nil {
			return nil, err
		}

		if trig != nil {
			return trig, nil
		}
	}

	return nil, nil
}

func BuildTriggersHandler(s *run.JobStore, rr run.Repo, p []Parser, logger logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		span, ctx := tracing.NewServiceSpan(ctx, "triggers")
		defer span.Finish()

		defer req.Body.Close()

		trig, err := parse(p, req)
		if err != nil {
			span.RecordError(err)
			logger.Warnf("failed to parse request: %v", err)
			respondErr(w, err)
			return
		}

		if trig == nil {
			logger.Warnf("no parser found to process request")
			respond(w, http.StatusOK, m{"details": "no parser to process request"})
			return
		}

		j, err := s.Fetch(trig.JobName)
		if err != nil {
			span.RecordError(err)
			logger.Warnf("failed to fetch job: %s - %v", trig.JobName, err)
			respondErr(w, &httpError{Message: err.Error(), Status: http.StatusNotFound})
			return
		}

		r := run.NewRun(j, *trig)
		if err = rr.CreateRun(context.TODO(), r); err != nil {
			span.RecordError(err)
			logger.Errorf("failed to create run with job %v, trigger; %v - %v", j, trig, err)
			respondErr(w, err)
			return
		}

		respond(w, http.StatusOK, r)
	}
}
