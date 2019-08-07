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

func (*LocalParser) Parse(r *http.Request) (*run.Trigger, error) {
	payload := localWebhook{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Parse request body")
	}

	if payload.Input == nil {
		payload.Input = make(map[string]interface{})
	}

	return &run.Trigger{
		JobName: payload.JobName,
		Scope:   payload.Scope,
		Input:   payload.Input,
	}, nil
}
