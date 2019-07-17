package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/pkg/errors"
)

type LocalParser struct{}

type localWebhook struct {
	JobName string                 `json:"job_name"`
	Scope   string                 `json:"scope"`
	Input   map[string]interface{} `json:"input_data"`
}

func (*LocalParser) Parse(r *http.Request) (run.Trigger, error) {
	payload := localWebhook{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&payload)
	if err != nil {
		return run.Trigger{}, errors.Wrapf(err, "failed to Parse request body")
	}

	return run.Trigger{
		JobName: payload.JobName,
		Scope:   payload.Scope,
		Input:   payload.Input,
	}, nil
}