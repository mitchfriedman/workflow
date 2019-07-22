package testhelpers_test

import (
	"github.com/mitchfriedman/workflow/lib/run"
)

func CreateStepperStore() *run.StepperStore {
	res := run.Result{State: run.StateSuccess, Data: make(run.InputData)}
	ss := run.NewStepperStore()
	ss.Register(NewSampleStep(res, "say_hello", nil))
	ss.Register(NewSampleStep(res, "say_goodbye", nil))
	ss.Register(NewSampleStep(res, "ask_question", nil))
	return ss
}

func CreateStep(t string) *run.Step {
	pID := generateUUID("ST")
	return &run.Step{
		UUID:     pID,
		StepType: t,
		State:    run.StateQueued,
		Input:    nil,
		Output:   run.Result{Data: make(map[string]interface{})},
	}
}
