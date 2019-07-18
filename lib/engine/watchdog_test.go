package engine_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mitchfriedman/workflow/lib/engine"
	"github.com/mitchfriedman/workflow/lib/run"
	testhelpers "github.com/mitchfriedman/workflow/lib/testhelpers"
	"github.com/mitchfriedman/workflow/lib/worker"
)

func TestWatchdog(t *testing.T) {
	db, closer := testhelpers.DBConnection(t, false)
	defer closer()

	rr := run.NewDatabaseStorage(db)
	wr := worker.NewDatabaseStorage(db)

	aWorkerID := "123"
	earlier := time.Now().AddDate(0, 0, -1)
	later := time.Now().Add(2 * time.Minute)

	oldInProgressRun := testhelpers.CreateSampleRun("job", "s1", make(run.InputData))
	oldInProgressRun.ClaimedBy = &aWorkerID
	oldInProgressRun.ClaimedUntil = &earlier

	currentWorker := worker.NewWorker()
	currentWorker.LastUpdated = earlier
	currentWorker.LeaseClaimedUntil = time.Now().Add(2 * time.Minute)

	currentInProgressRun := testhelpers.CreateSampleRun("job", "s2", make(run.InputData))
	currentInProgressRun.ClaimedBy = &currentWorker.UUID
	currentInProgressRun.ClaimedUntil = &later

	oldWorker := worker.NewWorker()
	oldWorker.LastUpdated = earlier
	oldWorker.LeaseClaimedUntil = time.Now().AddDate(0, 0, -1)

	tests := map[string]struct {
		runsBefore         []*run.Run
		workersBefore      []*worker.Worker
		runsAfterClaimed   []*run.Run
		runsAfterUnclaimed []*run.Run
		workersAfter       []*worker.Worker
	}{
		"with no existing workers or runs":                               {[]*run.Run{}, []*worker.Worker{}, []*run.Run{}, []*run.Run{}, []*worker.Worker{}},
		"with runs to release":                                           {[]*run.Run{oldInProgressRun}, []*worker.Worker{}, []*run.Run{}, []*run.Run{oldInProgressRun}, []*worker.Worker{}},
		"with runs to leave and release":                                 {[]*run.Run{oldInProgressRun, currentInProgressRun}, []*worker.Worker{currentWorker}, []*run.Run{currentInProgressRun}, []*run.Run{oldInProgressRun}, []*worker.Worker{currentWorker}},
		"with workers to remove":                                         {[]*run.Run{}, []*worker.Worker{oldWorker}, []*run.Run{}, []*run.Run{}, []*worker.Worker{}},
		"with workers to leave":                                          {[]*run.Run{}, []*worker.Worker{currentWorker}, []*run.Run{}, []*run.Run{}, []*worker.Worker{currentWorker}},
		"with workers to leave and remove and runs to leave and release": {[]*run.Run{oldInProgressRun, currentInProgressRun}, []*worker.Worker{currentWorker, oldWorker}, []*run.Run{currentInProgressRun}, []*run.Run{oldInProgressRun}, []*worker.Worker{currentWorker}},
	}

	cleanup := func() {
		var runs []*run.Run
		assert.Nil(t, db.Master.Find(&runs).Error)
		assert.Nil(t, db.Master.Delete(&runs).Error)

		var workers []*worker.Worker
		assert.Nil(t, db.Master.Find(&workers).Error)
		assert.Nil(t, db.Master.Delete(&workers).Error)
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cleanup()
			for _, d := range tc.runsBefore {
				assert.Nil(t, d.MarshalRunData())
				assert.Nil(t, db.Master.Create(&d).Error)
			}
			for _, d := range tc.workersBefore {
				assert.Nil(t, db.Master.Create(&d).Error)
			}

			engine.Process(context.Background(), log.New(os.Stderr, "", log.LstdFlags), wr, rr)
			var allRuns []*run.Run
			var allWorkers []*worker.Worker

			assert.Nil(t, db.Master.Find(&allRuns).Error)
			assert.Nil(t, db.Master.Find(&allWorkers).Error)

			verifyRuns(t, allRuns, tc.runsAfterClaimed, tc.runsAfterUnclaimed)
			verifyWorkers(t, allWorkers, tc.workersAfter)
		})
	}
}

func verifyWorkers(t *testing.T, actual []*worker.Worker, expected []*worker.Worker) {
	t.Helper()

	assert.Equal(t, len(expected), len(actual))
	expectedIds := make(map[string]struct{})
	for _, r := range expected {
		expectedIds[r.UUID] = struct{}{}
	}

	for _, r := range actual {
		_, ok := expectedIds[r.UUID]
		if !ok {
			t.Error(t, "run was found but not expected: "+r.UUID)
		}
	}
}

func verifyRuns(t *testing.T, actual []*run.Run, runsClaimed []*run.Run, runsUnclaimed []*run.Run) {
	t.Helper()

	assert.Equal(t, len(runsClaimed)+len(runsUnclaimed), len(actual))
	actualIds := make(map[string]run.Run)
	for _, r := range actual {
		actualIds[r.UUID] = *r
	}

	for _, r := range runsClaimed {
		// verify we have it actually stored in the db.
		actual, ok := actualIds[r.UUID]
		if !ok {
			t.Error(t, "run was found but not expected: "+r.UUID)
		}

		// verify it's claimed still.
		assert.NotNil(t, actual.ClaimedUntil)
		assert.NotNil(t, actual.ClaimedBy)
	}

	for _, r := range runsUnclaimed {
		// verify we have it actually stored in the db.
		actual, ok := actualIds[r.UUID]
		if !ok {
			t.Error(t, "run was found but not expected: "+r.UUID)
		}

		// verify it's not claimed anymore.
		assert.Nil(t, actual.ClaimedUntil)
		assert.Nil(t, actual.ClaimedBy)
	}
}
