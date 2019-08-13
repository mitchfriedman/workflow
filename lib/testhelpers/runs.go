package testhelpers_test

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mitchfriedman/workflow/lib/run"
)

func CreateSampleRunFirst2StepsSuccess(jobName, scope string, input run.InputData) *run.Run {
	r := CreateSampleRunFirstStepSuccess(jobName, scope, input)
	r.Steps.OnSuccess.State = run.StateSuccess
	return r
}

func CreateSampleRunFirstStepSuccess(jobName, scope string, input run.InputData) *run.Run {
	r := CreateSampleRun(jobName, scope, input)
	r.Steps.State = run.StateSuccess // complete the first step
	return r
}

func CreateSampleRunFirstStepFailureThenSuccess(jobName, scope string, input run.InputData) *run.Run {
	r := CreateSampleRunFirstStepFailure(jobName, scope, input)
	r.Steps.OnFailure.State = run.StateSuccess
	return r
}

func CreateSampleRunFirstStepFailure(jobName, scope string, input run.InputData) *run.Run {
	r := CreateSampleRun(jobName, scope, input)
	r.Steps.State = run.StateFailed // complete the first step
	return r
}

func CreateSampleRun(jobName, scope string, input run.InputData) *run.Run {
	/*

			CreateRun a graph that has a few paths. It looks like:

				       goodbye
			          / (success)
			       ask
			     /    \ (failed)
			hello      goodbye
		       \			  goodbye
				\ (failed) / (success)
				  goodbye

	*/
	hello := CreateStep("say_hello")
	ask := CreateStep("ask_question")
	goodbye := CreateStep("say_goodbye")
	goodbye2 := CreateStep("say_goodbye")
	hello.OnSuccess = ask
	ask.OnSuccess = goodbye
	ask.OnFailure = goodbye
	hello.OnFailure = goodbye
	hello.OnFailure.OnSuccess = goodbye2

	j := run.NewJob(jobName, hello)
	trig := run.Trigger{
		JobName: jobName,
		Scope:   scope,
		Input:   input,
	}

	return run.NewRun(j, trig)
}

func CreateSampleResultWithOutput(state run.State, ks ...string) run.Result {
	m := make(map[string]interface{}, len(ks)/2)
	for i := 0; i < len(ks); i += 2 {
		m[ks[i]] = ks[i+1]
	}
	return run.Result{
		State: state,
		Data:  m,
	}
}

func generateUUID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, string(uuid.New().String()))
}
