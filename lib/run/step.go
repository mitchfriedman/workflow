package run

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

type State string

const (
	StateQueued  State = "queued"
	StateSuccess State = "success"
	StateFailed  State = "failed"
	StateError   State = "error"
)

type Step struct {
	UUID      string    `json:"uuid"`
	Input     InputData `json:"input"`
	OnFailure *Step     `json:"on_failure"`
	OnSuccess *Step     `json:"on_success"`
	Output    Result    `json:"output"`
	State     State     `json:"state"`
	StepType  string    `json:"step_type"`
}

var ErrMissingRequiredInput = errors.New("required input is missing")

func InputSatisfied(data InputData, requiredInput []Input) error {
	for _, ri := range requiredInput {
		_, ok := data[ri.Name]
		if !ok {
			return errors.Wrapf(ErrMissingRequiredInput, "missing field %+v", ri)
		}
		// TODO: validate the type of the input.
	}

	return nil
}

func (s *Step) Terminal() bool {
	return s.OnSuccess == nil && s.OnFailure == nil
}

type Result struct {
	State State     `json:"state"`
	Data  InputData `json:"data"`
	Error string    `json:"error"`
}

func generateGraphFromStepTemplate(s *Step) *Step {
	if s == nil {
		return nil
	}

	step := stepFactory(s)
	step.OnSuccess = generateGraphFromStepTemplate(s.OnSuccess)
	step.OnFailure = generateGraphFromStepTemplate(s.OnFailure)

	return step
}

func generateUUID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, string(uuid.New().String()))
}

func stepFactory(t *Step) *Step {
	pID := generateUUID("ST")
	return &Step{
		Input:    t.Input,
		State:    StateQueued,
		StepType: t.StepType,
		UUID:     pID,
		Output:   Result{Data: make(map[string]interface{})},
	}
}
