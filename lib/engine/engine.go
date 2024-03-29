package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mitchfriedman/workflow/lib/tracing"

	"github.com/DataDog/datadog-go/statsd"

	"github.com/mitchfriedman/workflow/lib/logging"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/worker"
)

/*
Engine is a processor of steps. It's algorithms is:

At startup and in the background, begin a heartbeat goroutine to update it's
TTL every X seconds to indicate it's still doing work.

Then, enter into an infinite loop to perform the following:

1. Exit if we should stop processing.
2. Poll for the next step to execute.
3. Execute the step.
4. Go back to 1.

*/

var defaultPollAfter = 5 * time.Second
var defaultLeaseDuration = time.Minute
var defaultLeaseRenewDuration = 15 * time.Second

type Engine struct {
	w          *worker.Worker
	ss         *run.StepperStore
	rr         run.Repo
	wr         worker.Repo
	logger     logging.StructuredLogger
	metrics    *statsd.Client
	heartbeats chan worker.Heartbeat

	leaseDuration      time.Duration
	leaseRenewDuration time.Duration
	pollAfter          time.Duration
}

type Option func(e *Engine)

func WithLeaseDuration(d time.Duration) Option {
	return func(e *Engine) {
		e.leaseDuration = d
	}
}

func WithLeaseRenewDuration(d time.Duration) Option {
	return func(e *Engine) {
		e.leaseRenewDuration = d
	}
}

func WithPollAfter(d time.Duration) Option {
	return func(e *Engine) {
		e.pollAfter = d
	}
}

func NewEngine(w *worker.Worker, ss *run.StepperStore, rr run.Repo, wr worker.Repo, heartbeats chan worker.Heartbeat, logger logging.StructuredLogger, metrics *statsd.Client, options ...Option) *Engine {
	e := &Engine{w: w, ss: ss, rr: rr, wr: wr, heartbeats: heartbeats, logger: logger}
	e.leaseDuration = defaultLeaseDuration
	e.leaseRenewDuration = defaultLeaseRenewDuration
	e.pollAfter = defaultPollAfter
	e.logger = logger
	e.metrics = metrics

	for _, opt := range options {
		opt(e)
	}
	return e
}

func (e *Engine) Start(ctx context.Context) (err error) {
	span, ctx := tracing.NewServiceSpan(ctx, "engine.start")
	defer func() {
		span.RecordError(err)
		span.Finish()
	}()
	go e.heartbeat(ctx)

	var terminate bool
	var termMu sync.RWMutex

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				termMu.Lock()
				terminate = true
				termMu.Unlock()
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}(ctx)

	for {
		termMu.Lock()
		if terminate {
			termMu.Unlock()
			break
		}
		termMu.Unlock()
		err := e.process(ctx)
		if err != nil {
			e.logger.Errorf("failed to process steps: %v", err)
		}
		time.Sleep(e.pollAfter)
	}

	return nil
}

func executionStatus(err error) string {
	switch {
	case err == ErrNoRuns:
		return "no_runs"
	case err != nil:
		return "failed"
	default:
		return "success"
	}
}

func (e *Engine) process(ctx context.Context) (err error) {
	span, ctx := tracing.NewServiceSpan(ctx, "engine.process")
	defer func() {
		span.RecordError(err)
		span.Finish()
	}()

	// TODO: set context timeout.
	ex := NewExecutor(e.w.UUID, e.rr, e.ss)
	err = ex.Execute(ctx)

	e.metrics.Count("workflow.engine.execute", 1, []string{
		fmt.Sprintf("status:%s", executionStatus(err)),
	}, 1.0)

	switch err {
	case ErrNoRuns:
		return nil
	default:
		return err
	}
}

func (e *Engine) heartbeat(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(e.leaseRenewDuration):
			e.heartbeats <- worker.Heartbeat{Worker: *e.w, LeaseDuration: e.leaseDuration}
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}
