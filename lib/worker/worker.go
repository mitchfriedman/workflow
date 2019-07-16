package worker

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Worker struct {
	UUID              string
	LastUpdated       time.Time
	LeaseClaimedUntil time.Time
}

const prefix = "WO"

func NewWorker() *Worker {
	return &Worker{UUID: fmt.Sprintf("%s-%s", prefix, uuid.New().String())}
}
