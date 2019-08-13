package testhelpers_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mitchfriedman/workflow/lib/logging"

	database "github.com/mitchfriedman/workflow/lib/db"
	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/worker"

	"github.com/stretchr/testify/assert"
)

func DBConnection(t *testing.T, logMode ...bool) (*database.DB, func()) {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost:5432/workflow?sslmode=disable"
	}
	db, err := database.Connect(dbURL, dbURL, false, logging.New("test", os.Stderr))
	if err != nil {
		panic(fmt.Errorf("Could not connect to DB: %v\n", err))
	}

	if len(logMode) > 0 && logMode[0] {
		db.Master.LogMode(true)
		db.Reader.LogMode(true)
	}

	return db, func() {
		// delete everything from the database.
		var allRuns []*run.Run
		assert.Nil(t, db.Master.Find(&allRuns).Error)
		assert.Nil(t, db.Master.Delete(&allRuns).Error)

		var allWorkers []*worker.Worker
		assert.Nil(t, db.Master.Find(&allWorkers).Error)
		assert.Nil(t, db.Master.Delete(&allWorkers).Error)

		assert.Nil(t, db.Close())
	}
}
