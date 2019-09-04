package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mitchfriedman/workflow/lib/engine"
	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/testhelpers"
)

func TestCalculateRunStateTransition(t *testing.T) {
	ns := run.Step{}

	tests := map[string]struct {
		resultState  run.State
		isRollback   bool
		onFailure    *run.Step
		onSuccess    *run.Step
		wantRunState run.State
		wantRollback bool
	}{
		"success, with next step, not rollback": {run.StateSuccess, false, nil, &ns, run.StateQueued, false},
		"success, with next step, rollback":     {run.StateSuccess, true, nil, &ns, run.StateQueued, true},
		"success, with no next, not rollback":   {run.StateSuccess, false, nil, nil, run.StateSuccess, false},
		"success, with no next, rollback":       {run.StateSuccess, true, nil, nil, run.StateFailed, true},
		"failed, with next, not rollback":       {run.StateFailed, false, &ns, nil, run.StateQueued, true},
		"failed, with next, rollback":           {run.StateFailed, true, &ns, nil, run.StateQueued, true},
		"failed, with no next, not rollback":    {run.StateFailed, false, nil, nil, run.StateError, false},
		"failed, with no next, rollback":        {run.StateFailed, true, nil, nil, run.StateError, true},
		"queued, rollback":                      {run.StateQueued, true, nil, nil, run.StateQueued, true},
		"queued, not rollback":                  {run.StateQueued, false, nil, nil, run.StateQueued, false},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			ars, arb := engine.CalculateRunStateTransition(tc.resultState, tc.isRollback, tc.onSuccess, tc.onFailure)
			assert.Equal(t, tc.wantRunState, ars)
			assert.Equal(t, tc.wantRollback, arb)
		})
	}
}

func TestExecutor(t *testing.T) {
	db, closer := testhelpers.DBConnection(t, false)
	defer closer()
	repo := run.NewDatabaseStorage(db)
	ss := testhelpers.CreateStepperStore()

	r := testhelpers.CreateSampleRun("job", "s1", make(run.InputData))
	r.Steps.OnSuccess.OnSuccess.OnSuccess = nil
	assert.Nil(t, repo.CreateRun(context.Background(), r))

	executor := engine.NewExecutor("123", repo, ss)

	getRun := func(runId string) *run.Run {
		var r run.Run
		assert.Nil(t, db.Master.Where("uuid = ?", runId).First(&r).Error)
		assert.NotNil(t, &r)
		assert.Nil(t, r.UnmarshalRunData())
		return &r
	}

	var timeBegin time.Time
	for i := 0; i < 3; i++ { // 3 steps in run.
		timeBegin = time.Now().UTC()
		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		r := getRun(r.UUID)
		assert.NotNil(t, r.LastStepComplete)
		assert.True(t, r.LastStepComplete.After(timeBegin))
	}

	// if we try to execute again, we shouldn't find any runs since this one is completed.
	err := executor.Execute(context.Background())
	assert.Equal(t, engine.ErrNoRuns, err)

	// ensure that all steps have input
	r = getRun(r.UUID)
	ensureStepsContainInput(t, r.Steps)
}

func ensureStepsContainInput(t *testing.T, s *run.Step) {
	t.Helper()

	// if it hasn't been run or it's the leaf node.
	if s == nil || s.State == run.StateQueued {
		return
	}

	assert.NotNil(t, s.Input)
	assert.NotEmpty(t, s.Input)
	assert.Contains(t, s.Input, "step_uuid")
	assert.Contains(t, s.Input, "run_uuid")
	ensureStepsContainInput(t, s.OnSuccess)
	ensureStepsContainInput(t, s.OnFailure)
}
