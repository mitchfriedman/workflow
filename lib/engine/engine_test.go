package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/mitchfriedman/workflow/lib/metrics"

	"github.com/mitchfriedman/workflow/lib/logging"

	database2 "github.com/mitchfriedman/workflow/lib/db"
	"github.com/mitchfriedman/workflow/lib/engine"
	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/worker"

	"github.com/stretchr/testify/assert"

	"github.com/mitchfriedman/workflow/lib/testhelpers"
)

func setupRun(t *testing.T, rr run.Repo) string {
	t.Helper()

	r := testhelpers.CreateSampleRun("job", "s1", make(run.InputData))
	assert.Nil(t, rr.CreateRun(context.Background(), r))
	return r.UUID
}

func testHeartbeatsIncoming(t *testing.T, ctx context.Context, hbs chan worker.Heartbeat, hbDuration time.Duration) {
	go func(ctx context.Context) {
		var lastHbTime time.Time
		var missed int
		for {
			select {
			case <-ctx.Done():
				return
			case <-hbs:
				lastHbTime = time.Now().UTC()
			case <-time.After(2 * hbDuration):
				var empty time.Time
				if lastHbTime == empty {
					continue
				}

				if time.Now().UTC().Sub(lastHbTime) > hbDuration {
					missed++
					if missed > 10 {
						t.Error("missed more than 10 heartbeats")
					}
				}
			}
		}
	}(ctx)
}

func setupEngine(t *testing.T, ss *run.StepperStore, db *database2.DB, hbs chan worker.Heartbeat, hbDuration time.Duration) (*engine.Engine, *run.Storage, *worker.Worker, string, context.Context, context.CancelFunc) {
	t.Helper()

	w := worker.NewWorker()

	rr := run.NewDatabaseStorage(db)
	wr := worker.NewDatabaseStorage(db)
	runId := setupRun(t, rr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	logger := logging.New("test", os.Stderr)
	stats, _ := metrics.LoadStatsd("localhost", "9999", "", []string{}, logger)

	return engine.NewEngine(w, ss, rr, wr, hbs, logger, stats,
		engine.WithPollAfter(10*time.Nanosecond),
		engine.WithLeaseRenewDuration(hbDuration),
		engine.WithLeaseDuration(1*time.Millisecond)), rr, w, runId, ctx, cancel

}

func TestEngine(t *testing.T) {
	as1 := testhelpers.CreateStepperStore()

	as2 := testhelpers.CreateStepperStore()
	as2.Register(testhelpers.NewSampleStep(run.Result{State: run.StateFailed, Data: make(run.InputData)}, "say_hello", nil))

	as3 := testhelpers.CreateStepperStore()
	as3.Register(testhelpers.NewSampleStep(run.Result{State: run.StateError, Data: make(run.InputData)}, "say_hello", nil))

	tests := map[string]struct {
		ss              *run.StepperStore
		finalState      run.State
		failureStepName string
	}{
		"success": {as1, run.StateSuccess, ""},
		"failure": {as2, run.StateFailed, "say_hello"},
		"error":   {as3, run.StateError, "say_hello"},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			db, closer := testhelpers.DBConnection(t, true)
			defer closer()

			getRun := func(runId string) *run.Run {
				var r run.Run
				assert.Nil(t, db.Master.Where("uuid = ?", runId).First(&r).Error)
				assert.NotNil(t, &r)
				return &r
			}

			hbs := make(chan worker.Heartbeat, 1)
			hbDuration := 1 * time.Nanosecond
			e, _, _, runId, ctx, cancel := setupEngine(t, tc.ss, db, hbs, hbDuration)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					default:
						r := getRun(runId)
						if r.Terminal() {
							cancel()
							return
						}
						time.Sleep(10 * time.Millisecond)
					}
				}
			}()

			assert.Nil(t, e.Start(ctx))
			testHeartbeatsIncoming(t, ctx, hbs, hbDuration)
			wg.Wait()

			r := getRun(runId)
			var rd run.Data
			err := json.Unmarshal([]byte(r.Data), &rd)
			if err != nil {
				fmt.Printf("got: %s\n", string(r.Data))
			}
			assert.Nil(t, err)
			r.Input = rd.Input
			r.Steps = rd.Steps

			assert.Equal(t, run.State(r.State), tc.finalState)
			ensureStepsStatus(t, r.Steps, tc.failureStepName, tc.finalState)
		})
	}
}

func ensureStepsStatus(t *testing.T, step *run.Step, finalStepName string, finalState run.State) {
	if step == nil {
		return
	}

	if step.StepType == finalStepName {
		assert.Equal(t, run.State(step.State), finalState)
	} else {
		assert.Contains(t, []run.State{run.StateSuccess, run.StateQueued}, run.State(step.State))
		ensureStepsStatus(t, step.OnSuccess, finalStepName, finalState)
		ensureStepsStatus(t, step.OnFailure, finalStepName, finalState)
	}
}
