package rest_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mitchfriedman/workflow/lib/logging"

	"github.com/mitchfriedman/workflow/lib/testhelpers"

	"github.com/stretchr/testify/assert"

	"github.com/mitchfriedman/workflow/lib/rest"

	"github.com/mitchfriedman/workflow/lib/run"
)

type fakeParser struct {
	trigger *run.Trigger
	err     error
}

func (f *fakeParser) Parse(*http.Request) (*run.Trigger, error) {
	return f.trigger, f.err
}
func TestTriggers(t *testing.T) {
	db, closer := testhelpers.DBConnection(t, false)
	defer closer()

	rr := run.NewDatabaseStorage(db)
	jobName := "job1"
	r1 := testhelpers.CreateSampleRun(jobName, "s1", make(run.InputData))
	js := run.NewJobsStore()
	js.Register(run.NewJob(jobName, r1.Steps))

	parser1 := &fakeParser{nil, nil}
	parser2 := &fakeParser{nil, errors.New("bad")}
	parser3 := &fakeParser{&run.Trigger{JobName: jobName}, nil}

	//router := rest.NewRouter("test", run.NewJobsStore(), ni, nil)
	tests := map[string]struct {
		parsers    []rest.Parser
		wantStatus int
	}{
		"with no trigger parsed":                       {[]rest.Parser{parser1}, http.StatusOK},
		"with a trigger error":                         {[]rest.Parser{parser2}, http.StatusInternalServerError},
		"with no trigger parsed then error":            {[]rest.Parser{parser1, parser2}, http.StatusInternalServerError},
		"with successful parse":                        {[]rest.Parser{parser3}, http.StatusOK},
		"with no trigger parsed then successful parse": {[]rest.Parser{parser1, parser3}, http.StatusOK},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			b := []byte(fmt.Sprintf(`{"job_name": %s}`, "job1"))
			router := rest.NewRouter("test", js, rr, tc.parsers, logging.New("test", os.Stderr))
			req := httptest.NewRequest(http.MethodPost, "/Triggers", bytes.NewBuffer(b))
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantStatus, resp.Code)
		})
	}
}
