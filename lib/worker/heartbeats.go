package worker

import (
	"context"
	"log"
	"time"
)

type Heartbeat struct {
	WorkerID      string
	LeaseDuration time.Duration
}

type HeartbeatProcessor struct {
	hb     chan Heartbeat
	l      Leaser
	logger log.Logger
}

func NewHeartbeatProcessor(hb chan Heartbeat, leaser Leaser) *HeartbeatProcessor {
	return &HeartbeatProcessor{
		hb: hb,
		l:  leaser,
	}
}

func (h *HeartbeatProcessor) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case hb, ok := <-h.hb:
				if !ok {
					return
				}
				err := h.l.RenewLease(ctx, hb.WorkerID, hb.LeaseDuration)
				if err != nil {
					h.logger.Printf("heartbeat: failed to renew lease: %v", err)
				}
			}
		}
	}()
}
