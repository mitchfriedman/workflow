package rest_test

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mitchfriedman/workflow/lib/logging"

	"github.com/mitchfriedman/workflow/lib/rest"
	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/testhelpers"

	"github.com/stretchr/testify/assert"
)

func TestGetRun(t *testing.T) {
	db, closer := testhelpers.DBConnection(t, false)
	defer closer()

	r1 := testhelpers.CreateSampleRun("job1", "s1", make(run.InputData))
	rr := run.NewDatabaseStorage(db)
	rr.CreateRun(context.Background(), r1)

	tests := map[string]struct {
		uuid       string
		wantRun    *run.Run
		wantStatus int
	}{
		"with run present":  {r1.UUID, r1, 200},
		"with no run found": {"other", nil, 404},
	}

	router := rest.NewRouter("test", run.NewJobsStore(), rr, nil, logging.New("test", os.Stderr))

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/Runs/"+tc.uuid, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantStatus, resp.Code)
			if tc.wantRun != nil {
				var result *rest.RunRepresentation
				resultFrom(t, &result, resp.Body)
				assert.Equal(t, tc.wantRun.UUID, result.UUID)
			}
		})
	}
}

func resultFrom(t *testing.T, result interface{}, r io.Reader) {
	t.Helper()

	body, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
}

func TestGetRuns(t *testing.T) {
	db, closer := testhelpers.DBConnection(t, false)
	defer closer()
	j1s1 := testhelpers.CreateSampleRun("job1", "s1", make(run.InputData))
	j2s1 := testhelpers.CreateSampleRun("job2", "s1", make(run.InputData))
	j1s2 := testhelpers.CreateSampleRun("job1", "s2", make(run.InputData))
	js1s12 := testhelpers.CreateSampleRun("job1", "s1", make(run.InputData))

	rr := run.NewDatabaseStorage(db)
	rr.CreateRun(context.Background(), j1s1)
	rr.CreateRun(context.Background(), j1s2)
	rr.CreateRun(context.Background(), j2s1)
	rr.CreateRun(context.Background(), js1s12)

	tests := map[string]struct {
		job            string
		expectedRuns   []*run.Run
		expectedStatus int
	}{
		"with jobs present":           {"job1", []*run.Run{j1s1, j2s1, js1s12}, 200},
		"with jobs query not present": {"", []*run.Run{}, 400},
		"with jobs none found":        {"mr shneebly", []*run.Run{}, 200},
	}
	router := rest.NewRouter("test", run.NewJobsStore(), rr, nil, logging.New("test", os.Stderr))

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

			router.ServeHTTP(resp, req)
			assert.Equal(t, tc.expectedStatus, resp.Code)
			result := struct {
				Runs []rest.RunRepresentation `json:"runs"`
			}{}
			resultFrom(t, &result, resp.Body)
			assert.Equal(t, len(tc.expectedRuns), len(result.Runs))
		})
	}
}
