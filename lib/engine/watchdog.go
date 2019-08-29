package engine

import (
	"context"
	"time"

	"github.com/mitchfriedman/workflow/lib/logging"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/worker"

	"github.com/pkg/errors"
)

var watchDuration = 10 * time.Second

func Watch(ctx context.Context, logger logging.StructuredLogger, wr worker.Repo, rr run.Repo, runExpiry time.Duration) {
	for {
		select {
		case <-ctx.Done():
			break
		case <-time.After(watchDuration):
			Process(ctx, logger, wr, rr, runExpiry)
		}
	}
}

func Process(ctx context.Context, logger logging.StructuredLogger, wr worker.Repo, rr run.Repo, runExpiry time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), watchDuration)
	defer cancel()

	if err := cleanupWorkers(ctx, wr); err != nil {
		logger.Printf("watchdog: failed to cleanup workers: %v", err)
		return
	}

	if err := cleanupRuns(ctx, rr, wr, runExpiry); err != nil {
		logger.Printf("watchdog: failed to cleanup runs: %v", err)
	}
}

func cleanupWorkers(ctx context.Context, wr worker.Repo) error {
	allWorkers, err := wr.List(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch all workers")
	}

	for _, w := range allWorkers {
		if w.LeaseClaimedUntil.UTC().After(time.Now().UTC()) {
			continue
		}

		if err := wr.Deregister(ctx, w.UUID); err != nil {
			return errors.Wrapf(err, "failed to deregister worker %v", w)
		}
	}

	return nil
}

func cleanupRuns(ctx context.Context, rr run.Repo, wr worker.Repo, runExpiry time.Duration) error {
	runs, err := rr.ClaimedRuns(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to fetch claimed runs")
	}

	for _, r := range runs {
		// if it's currently unclaimed, there's nothing to do for this run.
		if r.ClaimedBy == nil && r.ClaimedUntil == nil {
			continue
		}

		// if the run is not making progress for longer than the expiry time, let's time it out.
		if r.LastStepComplete != nil && time.Now().Sub(*r.LastStepComplete) > runExpiry {
			r.State = run.StateError
			if err := rr.ReleaseRun(ctx, r); err != nil {
				return errors.Wrapf(err, "cleanupRuns: failed to abort and release run: %v", r)
			}
			continue
		}

		// fetch this worker and see if it is still around.
		w, err := wr.Get(ctx, *r.ClaimedBy)
		if err != nil {
			return errors.Wrapf(err, "cleanupRuns: failed to get run: %v", r)
		}

		// the worker is present.
		if w != nil {
			continue
		}

		// the worker is no longer with us, let's release this run and let another claim it.
		if err := rr.ReleaseRun(ctx, r); err != nil {
			return errors.Wrapf(err, "cleanupRuns: failed to release run: %v", r)
		}
	}

	return nil
}
