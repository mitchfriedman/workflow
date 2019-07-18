package worker

import (
	"context"
	"time"

	database "github.com/mitchfriedman/workflow/lib/db"
)

type Repo interface {
	Leaser
	Registerer
	Lister
}

type Leaser interface {
	RenewLease(context.Context, *Worker, time.Duration) error
}

type Registerer interface {
	Register(context.Context, *Worker) error
	Deregister(context.Context, string) error
}

type Lister interface {
	List(context.Context) ([]*Worker, error)
}

type DatabaseStorage struct {
	db *database.DB
}

func NewDatabaseStorage(db *database.DB) *DatabaseStorage {
	return &DatabaseStorage{db: db}
}

func (d *DatabaseStorage) RenewLease(ctx context.Context, w *Worker, t time.Duration) error {
	w.LastUpdated = time.Now()
	w.LeaseClaimedUntil = time.Now().Add(t)
	return d.db.Master.Model(&w).Where("uuid = ?", w.UUID).Error
}

func (d *DatabaseStorage) Register(ctx context.Context, w *Worker) error {
	return d.db.Master.Create(w).Error
}

func (d *DatabaseStorage) Deregister(ctx context.Context, workerId string) error {
	w := Worker{UUID: workerId}
	return d.db.Master.Where("uuid = ?", workerId).Delete(&w).Error
}

func (d *DatabaseStorage) List(ctx context.Context) ([]*Worker, error) {
	var all []*Worker
	err := d.db.Reader.Find(&all).Error
	return all, err
}
