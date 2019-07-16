package run

import (
	"context"
	"time"

	database "github.com/mitchfriedman/workflow/lib/db"
	"github.com/pkg/errors"
)

type Repo interface {
	Creator
	Retriever
	Claimer
}

type Retriever interface {
	NextRuns() ([]*Run, error)
	ClaimedRuns(context.Context) ([]*Run, error)
	ListByJob(context.Context, string) ([]*Run, error)
}

type Claimer interface {
	Claim(*Run, string, time.Duration) error
	Release(*Run) error
}

type Creator interface {
	Create(*Run) error
}

type Storage struct {
	db *database.DB
}

func NewDatabaseStorage(db *database.DB) *Storage {
	return &Storage{db: db}
}

func (r *Storage) Create(d *Run) error {
	err := d.MarshalRunData()
	if err != nil {
		return err
	}
	d.Started = time.Now()

	return r.db.Master.Create(&d).Error
}

func (r *Storage) NextRuns() ([]*Run, error) {
	var runs []*Run
	err := r.db.Reader.
		Where("state = ?", StateQueued).
		Find(&runs).Error

	if err != nil {
		return nil, errors.Wrap(err, "failed to query for next run")
	}

	for _, r := range runs {
		err := r.UnmarshalRunData()
		if err != nil {
			return nil, err
		}
	}

	return runs, nil
}

func (r *Storage) Claim(t *Run, workerID string, d time.Duration) error {
	err := t.MarshalRunData()
	if err != nil {
		return err
	}

	n := time.Now().Add(d)
	t.ClaimedBy = &workerID
	t.ClaimedUntil = &n

	return r.db.Master.
		Model(&t).
		Where("uuid = ?", t.UUID).
		Updates(t).Error
}

func (r *Storage) Release(d *Run) error {
	err := d.MarshalRunData()
	if err != nil {
		return err
	}

	n := time.Now()
	d.ClaimedBy = nil
	d.ClaimedUntil = nil
	d.LastStepComplete = &n

	updates := map[string]interface{}{
		"claimed_by":         nil,
		"claimed_until":      nil,
		"last_step_complete": &n,
		"data":               d.Data,
		"state":              d.State,
		"rollback":           d.Rollback,
	}

	return r.db.Master.
		Model(&d).
		Where("uuid = ?", d.UUID).
		Updates(updates).Error
}

func (r *Storage) ClaimedRuns(ctx context.Context) ([]*Run, error) {
	var runs []*Run
	err := r.db.Reader.
		Where("state = ?", StateQueued).
		Find(&runs).Error

	if err != nil {
		return nil, errors.Wrap(err, "failed to query runs")
	}

	return runs, nil
}

func (r *Storage) ListByJob(ctx context.Context, job string) ([]*Run, error) {
	var runs []*Run
	err := r.db.Reader.Where("job_name = ?", job).Find(&runs).Error
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query runs by job: %s", job)
	}

	return runs, nil
}
