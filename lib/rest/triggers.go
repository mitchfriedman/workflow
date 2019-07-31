package rest

import (
	"net/http"

	"github.com/mitchfriedman/workflow/lib/run"
)

func BuildTriggersHandler(s *run.JobStore, rr run.Repo, p Parser) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		trig, err := p.Parse(req)
		if err != nil {
			respondErr(w, err)
			return
		}

		j, err := s.Fetch(trig.JobName)
		if err != nil {
			respondErr(w, &httpError{Message: err.Error(), Status: http.StatusNotFound})
		}

		r := run.NewRun(j, trig)
		if err = rr.Create(r); err != nil {
			respondErr(w, err)
		}

		respond(w, http.StatusOK, r)
	}
}
