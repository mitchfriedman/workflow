package engine_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mitchfriedman/workflow/lib/engine"
	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/stretchr/testify/assert"

	testhelpers "github.com/mitchfriedman/workflow/lib/testhelpers"
)

func TestPrioritize(t *testing.T) {
	j1s1 := testhelpers.CreateSampleRun("job", "s1", make(run.InputData))
	j2s1 := testhelpers.CreateSampleRun("job", "s1", make(run.InputData))
	j3s1 := testhelpers.CreateSampleRun("job", "s1", make(run.InputData))
	j4s1 := testhelpers.CreateSampleRun("job", "s1", make(run.InputData))
	n := time.Now().UTC()
	later := time.Now().UTC().Add(10 * time.Second)
	workerId := "123"
	j3s1.LastStepComplete = &n
	j4s1.LastStepComplete = &n
	j4s1.ClaimedUntil = &later
	j4s1.ClaimedBy = &workerId

	tests := map[string]struct {
		runs        []*run.Run
		expectedRun *string
	}{
		"only 1 of key run+scope, in-progress, picks that one":          {runs: []*run.Run{j3s1}, expectedRun: &j3s1.UUID},
		"only 1 of key run+scope, not started, picks that one":          {runs: []*run.Run{j1s1}, expectedRun: &j1s1.UUID},
		"multiple of same run+scope, none started":                      {runs: []*run.Run{j1s1, j2s1}, expectedRun: &j1s1.UUID},
		"multiple of same run+scope, 1 already started but not claimed": {runs: []*run.Run{j1s1, j3s1}, expectedRun: &j3s1.UUID},
		"multiple of same run+scope, 1 claimed being executed":          {runs: []*run.Run{j1s1, j4s1}, expectedRun: nil},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			db, closer := testhelpers.DBConnection(t, false)
			defer closer()
			rr := run.NewDatabaseStorage(db)

			for _, r := range tc.runs {
				assert.Nil(t, rr.CreateRun(context.Background(), r))
			}

			p, err := engine.Prioritize(context.Background(), rr)
			assert.Nil(t, err)
			if tc.expectedRun == nil {
				assert.Nil(t, p)
			} else if p.UUID != *tc.expectedRun {
				t.Error(fmt.Sprintf("prioritized incorrect run: actual: %v - expected: %v", p, tc.expectedRun))
			}
		})
	}

}
