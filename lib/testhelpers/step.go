package testhelpers

import (
	"github.com/mitchfriedman/workflow/lib/run"
)

type SampleStep struct {
	Res run.Result
	Err error
	t   string
}

func NewSampleStep(res run.Result, t string, err error) run.Stepper {
	return &SampleStep{res, err, t}
}

func (p *SampleStep) Type() string {
	return p.t
}

func (p *SampleStep) RequiredInput() []run.Input {
	return []run.Input{}
}

func (p *SampleStep) Step(d run.InputData) (run.Result, error) {
	return p.Res, p.Err
}
