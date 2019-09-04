package run_test

import (
	"testing"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/testhelpers"

	"github.com/stretchr/testify/assert"
)

func TestSatisfied(t *testing.T) {
	tests := map[string]struct {
		input         run.InputData
		requiredInput []run.Input
		shouldSatisfy bool
	}{
		"satisfied with no input":                    {make(run.InputData), []run.Input{}, true},
		"satisfied with 1 input":                     {map[string]interface{}{"foo": "bar"}, []run.Input{{Name: "foo", Type: run.InputTypeString}}, true},
		"satisfied with 2 input":                     {map[string]interface{}{"foo": "bar", "foo2": "bar2"}, []run.Input{{Name: "foo", Type: run.InputTypeString}, {Name: "foo2", Type: run.InputTypeString}}, true},
		"not satisfied with missing input":           {map[string]interface{}{"foo": "bar"}, []run.Input{{Name: "foo2", Type: run.InputTypeString}}, false},
		"not satisfied with multiple missing inputs": {map[string]interface{}{"foo": "bar"}, []run.Input{{Name: "foo2", Type: run.InputTypeString}, {Name: "foo3", Type: run.InputTypeString}}, false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.shouldSatisfy {
				assert.Nil(t, run.InputSatisfied(tc.input, tc.requiredInput))
			} else {
				assert.NotNil(t, run.InputSatisfied(tc.input, tc.requiredInput))
			}
		})
	}
}

func TestNextStep(t *testing.T) {
	jobInput := map[string]interface{}{
		"foo": "bar",
	}
	r1Input := map[string]interface{}{
		"foo": "bar",
	}
	r2Input1 := map[string]interface{}{
		"foo":  "bar",
		"foo2": "bar2",
	}
	r2Input2 := map[string]interface{}{
		"foo":  "bar",
		"foo2": "bar2",
	}
	r3Input1 := map[string]interface{}{
		"foo":  "bar",
		"foo2": "bar2",
		"foo3": "bar3",
	}
	r3Input2 := map[string]interface{}{
		"foo":  "bar",
		"foo2": "bar2",
		"foo3": "bar3",
	}
	r1 := testhelpers.CreateSampleRun("job", "s1", jobInput)
	r1Input["run_uuid"] = r1.UUID
	r1Input["step_uuid"] = r1.Steps.UUID
	s1Output := testhelpers.CreateSampleResultWithOutput(run.StateSuccess, "foo2", "bar2")
	s2Output := testhelpers.CreateSampleResultWithOutput(run.StateSuccess, "foo3", "bar3")

	r2 := testhelpers.CreateSampleRunFirstStepSuccess("job", "s1", jobInput)
	r2.Steps.Output = s1Output
	r2Input1["run_uuid"] = r2.UUID
	r2Input1["step_uuid"] = r2.Steps.OnSuccess.UUID

	r3 := testhelpers.CreateSampleRunFirst2StepsSuccess("job", "s1", jobInput)
	r3.Steps.Output = s1Output
	r3.Steps.OnSuccess.Output = s2Output
	r3Input1["run_uuid"] = r3.UUID
	r3Input1["step_uuid"] = r3.Steps.OnSuccess.OnSuccess.UUID

	r4 := testhelpers.CreateSampleRunFirstStepFailure("job", "s1", jobInput)
	r4.Steps.Output = s1Output
	r2Input2["run_uuid"] = r4.UUID
	r2Input2["step_uuid"] = r4.Steps.OnFailure.UUID

	r5 := testhelpers.CreateSampleRunFirstStepFailureThenSuccess("job", "s1", jobInput)
	r5.Steps.Output = s1Output
	r5.Steps.OnFailure.Output = s2Output
	r3Input2["run_uuid"] = r5.UUID
	r3Input2["step_uuid"] = r5.Steps.OnFailure.OnSuccess.UUID

	r6 := testhelpers.CreateSampleRun("job", "s1", jobInput)
	r6.Steps.Output.Data["some_key"] = "some_value"
	r6Input := map[string]interface{}{
		"foo":       "bar",
		"run_uuid":  r6.UUID,
		"step_uuid": r6.Steps.UUID,
		"some_key":  "some_value",
	}

	tests := map[string]struct {
		run    *run.Run
		action string
		input  run.InputData
		err    error
	}{
		"run not started":                             {run: r1, action: r1.Steps.StepType, input: r1Input},
		"run started, on second, success":             {run: r2, action: r2.Steps.OnSuccess.StepType, input: r2Input1},
		"run started, on third, both success":         {run: r3, action: r3.Steps.OnSuccess.OnSuccess.StepType, input: r3Input1},
		"run started, on second, failure":             {run: r4, action: r4.Steps.OnFailure.StepType, input: r2Input2},
		"run started, on third, failure then success": {run: r5, action: r5.Steps.OnFailure.OnSuccess.StepType, input: r3Input2},
		"run started, already ran, uses output":       {run: r6, action: r6.Steps.StepType, input: r6Input},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			s, data, err := tc.run.NextStep()
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.action, s.StepType)
			assert.Equal(t, tc.input, data)
		})
	}
}
