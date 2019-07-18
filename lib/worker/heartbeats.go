package worker

import (
	"context"
	"log"
	"time"
)

type Heartbeat struct {
	Worker        Worker
	LeaseDuration time.Duration
}

type HeartbeatProcessor struct {
	hb     chan Heartbeat
	l      Leaser
	logger log.Logger
}

func NewHeartbeatProcessor(hb chan Heartbeat, l Leaser) *HeartbeatProcessor {
	return &HeartbeatProcessor{
		hb: hb,
		l:  l,
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
				h.logger.Printf("heartbeat: failed to renew lease: %v", err)
			}
		}
	}
}
