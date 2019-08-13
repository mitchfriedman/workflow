package rest

import (
	"context"
	"net/http"

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

func BuildTriggersHandler(s *run.JobStore, rr run.Repo, p []Parser) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		trig, err := parse(p, req)
		if err != nil {
			respondErr(w, err)
			return
		}

		if trig == nil {
			respondErr(w, &httpError{
				Message: "No parser found to process trigger",
				Status:  http.StatusBadRequest,
			})
			return
		}

		j, err := s.Fetch(trig.JobName)
		if err != nil {
			respondErr(w, &httpError{Message: err.Error(), Status: http.StatusNotFound})
			return
		}

		r := run.NewRun(j, *trig)
		if err = rr.CreateRun(context.TODO(), r); err != nil {
			respondErr(w, err)
			return
		}

		respond(w, http.StatusOK, r)
	}
}
