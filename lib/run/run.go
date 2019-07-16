package run

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

var ErrNoQueuedSteps = errors.New("failed to find queued step")

// Trigger is something that kicks off a Run.
type Trigger struct {
	JobName string
	Scope   string
	Input   InputData // input from the trigger source (API data, webhook, etc).
}

// Run is an instantiation of a Job.
type Run struct {
	Input InputData       `sql:"-"`
	Job   Job             `sql:"-"`
	Steps *Step           `sql:"-"`
	Data  json.RawMessage `gorm:"type:jsonb;"`

	JobName  string
	Rollback bool
	Scope    string // i.e. the application name
	State    State
	UUID     string

	Started          time.Time
	Finished         *time.Time
	LastStepComplete *time.Time
	ClaimedUntil     *time.Time
	ClaimedBy        *string // uuid of worker, if claimed
}

func (r *Run) MarshalRunData() error {
	rd := Data{
		Input: r.Input,
		Steps: r.Steps,
		Job:   r.Job,
	}
	var err error
	r.Data, err = json.Marshal(&rd)
	return err
}

func (r *Run) UnmarshalRunData() error {
	var rd Data
	err := json.Unmarshal(r.Data, &rd)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal r data")
	}
	r.Input = rd.Input
	r.Steps = rd.Steps
	r.Job = rd.Job

	return nil
}

func (r *Run) Terminal() bool {
	return r.State == StateFailed || r.State == StateSuccess || r.State == StateError
}

type Data struct {
	Input InputData `json:"input"`
	Steps *Step     `json:"step"`
	Job   Job       `json:"job"`
}

func NewRun(j Job, trigger Trigger) *Run {
	id := generateUUID("RU")
	steps := generateGraph(j.Start)
	steps.State = StateQueued

	return &Run{
		Input:   trigger.Input,
		JobName: j.Name,
		Job:     j,
		Scope:   trigger.Scope,
		State:   StateQueued,
		UUID:    id,
		Steps:   steps,
	}
}

func (r *Run) NextStep() (*Step, InputData, error) {
	firstQueued, data, err := findFirstQueuedStepAndHydrateInput(r.Steps, r.Input)
	if err != nil {
		return nil, nil, err
	}

	// inject relevant data before this step is executed.
	data["step_uuid"] = firstQueued.UUID
	data["run_uuid"] = r.UUID

	return firstQueued, data, nil
}

func findFirstQueuedStepAndHydrateInput(s *Step, d InputData) (*Step, InputData, error) {
	if s == nil {
		return nil, nil, nil
	}

	if s.State == StateQueued {
		return s, d, nil
	}

	if s.State == StateSuccess {
		return findFirstQueuedStepAndHydrateInput(s.OnSuccess, d.Merge(s.Output.Data))
	}

	if s.State == StateFailed {
		return findFirstQueuedStepAndHydrateInput(s.OnFailure, d.Merge(s.Output.Data))
	}

	return nil, nil, ErrNoQueuedSteps
}
