package engine

import (
	"context"
	"time"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/pkg/errors"
)

var claimDuration = 30 * time.Second

var ErrNoRuns = errors.New("no runs to execute")

type Executor struct {
	workerID     string
	runRepo      run.Repo
	stepperStore *run.StepperStore
}

func NewExecutor(workerID string, runRepo run.Repo, ss *run.StepperStore) *Executor {
	return &Executor{
		workerID:     workerID,
		runRepo:      runRepo,
		stepperStore: ss,
	}
}

func (p *Executor) Execute(ctx context.Context) error {
	r, err := p.nextRun()
	if err != nil {
		return err
	}

	if r == nil {
		return ErrNoRuns
	}

	err = p.claimRun(r)
	if err != nil {
		return err
	}

	s, input, err := r.NextStep()
	if err != nil {
		return err
	}

	stepper, err := p.getStepper(s)
	if err != nil {
		return err
	}

	if err := run.InputSatisfied(input, stepper.RequiredInput()); err != nil {
		err2 := p.abortRun(r)
		if err2 != nil {
			return errors.Wrap(err2, "failed trying to abort run")
		}
		return errors.Wrap(err, "step is not satisfied with input")
	}

	result, err := stepper.Step(input)
	if err != nil {
		return err
	}

	return p.updateAndReleaseRun(result, r, s)
}

func (p *Executor) abortRun(r *run.Run) error {
	r.State = run.StateError
	return p.runRepo.Release(r)
}

func (p *Executor) updateAndReleaseRun(result run.Result, r *run.Run, s *run.Step) error {
	// Update the step state and save the output to the step.
	// We want to do this even if the state hasn't changed because
	// the step might put useful information into that output that
	// can be used on future iterations.
	s.State = result.State
	s.Output = result

	r.State, r.Rollback = CalculateRunStateTransition(result.State, r.Rollback, s.OnSuccess, s.OnFailure)

	return p.runRepo.Release(r)
}

func CalculateRunStateTransition(resultState run.State, isRollback bool, onFailure, onSuccess *run.Step) (run.State, bool) {
	/* The state transition logic can be described as follows:
	1. step success?
		-> have an OnSuccess?
			-> execute the on success (run state -> queued)
		-> don't have an onSuccess?
			-> is this executing a current rollback?
				-> yes?
					-> run state is now failed (since it was a rollback)
				-> no?
					-> run state is now success (since not a rollback)
	2. step failure?
		-> have an OnFailure?
			-> run state is now queued, rollback = true
		-> don't have an OnFailure?
			-> run state is now error (no way to handle this failure)
	3. step errpr?
		-> go to error state, do not attempt rollback.
	*/
	switch resultState {
	case run.StateFailed:
		if onFailure == nil {
			return run.StateError, isRollback
		}
		return run.StateQueued, true
	case run.StateSuccess:
		if onSuccess == nil {
			if isRollback {
				return run.StateFailed, isRollback
			}
			return run.StateSuccess, isRollback
		} else {
			return run.StateQueued, isRollback
		}
	case run.StateError:
		return run.StateError, isRollback
	}
	// state is still queued - no change to the state or rollback status.
	return run.StateQueued, isRollback
}

func (p *Executor) getStepper(s *run.Step) (run.Stepper, error) {
	return p.stepperStore.Get(s.StepType)
}

func (p *Executor) claimRun(r *run.Run) error {
	return p.runRepo.Claim(r, p.workerID, claimDuration)
}

func (p *Executor) nextRun() (*run.Run, error) {
	return Prioritize(p.runRepo)
}
