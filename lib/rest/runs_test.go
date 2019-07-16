package rest_test

import (
	"encoding/json"
	"github.com/mitchfriedman/workflow/lib/rest"
	"github.com/mitchfriedman/workflow/lib/run"
	testhelpers "github.com/mitchfriedman/workflow/lib/testhelpers"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRuns(t *testing.T) {
	db, closer := testhelpers.DBConnection(t, false)
	defer closer()
	j1s1 := testhelpers.CreateSampleRun("job1", "s1", make(run.InputData))
	j2s1 := testhelpers.CreateSampleRun("job2", "s1", make(run.InputData))
	j1s2 := testhelpers.CreateSampleRun("job1", "s2", make(run.InputData))
	js1s12 := testhelpers.CreateSampleRun("job1", "s1", make(run.InputData))

	rr := run.NewDatabaseStorage(db)
	rr.Create(j1s1)
	rr.Create(j1s2)
	rr.Create(j2s1)
	rr.Create(js1s12)

	tests := map[string]struct {
		job            string
		expectedRuns   []*run.Run
		expectedStatus int
	}{
		"with jobs present":           {"job1", []*run.Run{j1s1, j2s1, js1s12}, 200},
		"with jobs query not present": {"", []*run.Run{}, 400},
		"with jobs none found":        {"mr shneebly", []*run.Run{}, 200},
	}
	handler := rest.BuildGetRunsHandler(rr)

	resultFrom := func(result interface{}, r io.Reader) {
		body, err := ioutil.ReadAll(r)
		assert.Nil(t, err)
		err = json.Unmarshal(body, &result)
		assert.Nil(t, err)
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/Runs", nil)
			resp := httptest.NewRecorder()
			if tc.job != "" {
				q := req.URL.Query()
				q.Add("job_name", tc.job)
				req.URL.RawQuery = q.Encode()
			}

			handler.ServeHTTP(resp, req)
			assert.Equal(t, tc.expectedStatus, resp.Code)
			result := struct {
				Runs []rest.RunRepresentation `json:"runs"`
			}{}
			resultFrom(&result, resp.Body)
			assert.Equal(t, len(tc.expectedRuns), len(result.Runs))
		})
	}
}
