package worker

import (
	"context"
	"time"

	"github.com/mitchfriedman/workflow/lib/tracing"

	"github.com/jinzhu/gorm"

	database "github.com/mitchfriedman/workflow/lib/db"
)

type Repo interface {
	Leaser
	Registerer
	Retriever
}

type Leaser interface {
	RenewLease(context.Context, *Worker, time.Duration) error
}

type Registerer interface {
	Register(context.Context, *Worker) error
	Deregister(context.Context, string) error
}

type Retriever interface {
	List(context.Context) ([]*Worker, error)
	Get(context.Context, string) (*Worker, error)
}

type DatabaseStorage struct {
	db *database.DB
}

func NewDatabaseStorage(db *database.DB) *DatabaseStorage {
	return &DatabaseStorage{db: db}
}

func (d *DatabaseStorage) RenewLease(ctx context.Context, w *Worker, t time.Duration) error {
	w.LastUpdated = time.Now().UTC()
	w.LeaseClaimedUntil = time.Now().UTC().Add(t)
	span, ctx := tracing.NewDBSpan(ctx, d.db.Master, "worker.renew_lease")
	err := d.db.Master.Model(&w).Where("uuid = ?", w.UUID).Update(&w).Error
	span.RecordError(err)
	span.Finish()

	return err
}

func (d *DatabaseStorage) Register(ctx context.Context, w *Worker) error {
	span, ctx := tracing.NewDBSpan(ctx, d.db.Master, "worker.register")
	err := d.db.Master.Create(w).Error
	span.RecordError(err)
	span.Finish()

	return err
}

func (d *DatabaseStorage) Deregister(ctx context.Context, workerId string) error {
	w := Worker{UUID: workerId}
	span, ctx := tracing.NewDBSpan(ctx, d.db.Master, "worker.deregister")
	err := d.db.Master.Where("uuid = ?", workerId).Delete(&w).Error
	span.RecordError(err)
	span.Finish()

	return err
}

func (d *DatabaseStorage) List(ctx context.Context) ([]*Worker, error) {
	var all []*Worker
	span, ctx := tracing.NewDBSpan(ctx, d.db.Master, "worker.list")
	err := d.db.Reader.Find(&all).Error
	span.RecordError(err)
	span.Finish()

	return all, err
}

func (d *DatabaseStorage) Get(ctx context.Context, uuid string) (*Worker, error) {
	var w Worker
	span, ctx := tracing.NewDBSpan(ctx, d.db.Master, "worker.get")
	err := d.db.Reader.Where("uuid = ?", uuid).First(&w).Error
	span.RecordError(err)
	span.Finish()

	switch err {
	case gorm.ErrRecordNotFound:
		return nil, nil
	}
	return &w, err
}
