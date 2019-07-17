package engine

import (
	"context"
	"log"
	"time"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/worker"

	"github.com/pkg/errors"
)

var watchDuration = 10 * time.Second

func Watch(ctx context.Context, logger *log.Logger, wr worker.Repo, rr run.Repo) {
	for {
		select {
		case <-ctx.Done():
			break
		case <-time.After(watchDuration):
			Process(ctx, logger, wr, rr)
		}
	}
}

func Process(ctx context.Context, logger *log.Logger, wr worker.Repo, rr run.Repo) {
	ctx, cancel := context.WithTimeout(context.Background(), watchDuration)
	defer cancel()

	if err := cleanupWorkers(ctx, wr); err != nil {
		logger.Printf("watchdog: failed to cleanup workers: %v", err)
		return
	}

	if err := cleanupRuns(ctx, rr); err != nil {
		logger.Printf("watchdog: failed to cleanup runs: %v", err)
	}
}

func cleanupWorkers(ctx context.Context, wr worker.Repo) error {
	allWorkers, err := wr.List(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch all workers")
	}

	for _, w := range allWorkers {
		if w.LeaseClaimedUntil.After(time.Now()) {
			continue
		}

		if err := wr.Deregister(ctx, w.UUID); err != nil {
			return errors.Wrapf(err, "failed to deregister worker %v", w)
		}
	}

	return nil
}

func cleanupRuns(ctx context.Context, rr run.Repo) error {
	runs, err := rr.ClaimedRuns(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch claimed runs")
	}

	for _, r := range runs {
		if r.ClaimedUntil == nil || (*r.ClaimedUntil).After(time.Now()) {
			continue
		}

		if err := rr.Release(r); err != nil {
			return errors.Wrapf(err, "failed to release run: %v", r)
		}
	}

	return nil
}
