package worker

import (
	"context"
	"time"

	"github.com/mitchfriedman/workflow/lib/logging"
)

type Heartbeat struct {
	Worker        Worker
	LeaseDuration time.Duration
}

type HeartbeatProcessor struct {
	hb     chan Heartbeat
	l      Leaser
	logger logging.StructuredLogger
}

func NewHeartbeatProcessor(hb chan Heartbeat, l Leaser, logger logging.StructuredLogger) *HeartbeatProcessor {
	return &HeartbeatProcessor{
		hb:     hb,
		l:      l,
		logger: logger,
	}
}

func (h *HeartbeatProcessor) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case hb, ok := <-h.hb:
			if !ok {
				return
			}
			err := h.l.RenewLease(ctx, &hb.Worker, hb.LeaseDuration)
			if err != nil {
				h.logger.Errorf("heartbeat: failed to renew lease: %v", err)
			}
		}
	}
}
