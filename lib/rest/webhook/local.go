package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/pkg/errors"
)

type LocalParser struct{}

type localWebhook struct {
	JobName string `json:"job_name"`
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
	}, nil
}
