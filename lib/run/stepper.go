package run

import "errors"

type Input struct {
	Name, Type string
}

type Stepper interface {
	Step(InputData) (Result, error)
	Type() string
	RequiredInput() []Input
}

type StepperStore struct {
	steppers map[string]Stepper
}

func NewStepperStore() *StepperStore {
	return &StepperStore{
		steppers: make(map[string]Stepper),
	}
}

func (s *StepperStore) Register(stepper Stepper) {
	s.steppers[stepper.Type()] = stepper
}

func (s *StepperStore) Get(t string) (Stepper, error) {
	stepper, ok := s.steppers[t]
	if !ok {
		return nil, errors.New("no such stepper found")
	}
	return stepper, nil
}
