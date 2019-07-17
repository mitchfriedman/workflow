package worker_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mitchfriedman/workflow/lib/worker"
)

type fakeLeaser struct {
	counter   int
	counterMu sync.Mutex
}

func (f *fakeLeaser) RenewLease(context.Context, string, time.Duration) error {
	f.counterMu.Lock()
	f.counter++
	f.counterMu.Unlock()
	return nil
}

func (f *fakeLeaser) count() int {
	f.counterMu.Lock()
	c := f.counter
	f.counterMu.Unlock()
	return c
}

func TestHeartbeatProcessor(t *testing.T) {
	leaser := &fakeLeaser{}
	hbs := make(chan worker.Heartbeat, 1)
	hbp := worker.NewHeartbeatProcessor(hbs, leaser)

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer wg.Done()
		hbp.Start(ctx)
	}()

	go func() {
		for i := 0; i < 10; i++ {
			hbs <- worker.Heartbeat{WorkerID: fmt.Sprintf("%d", i), LeaseDuration: time.Millisecond}
		}
	}()

	for {
		if leaser.count() == 10 {
			break
		}
	}

	cancel()
	wg.Wait()
}
