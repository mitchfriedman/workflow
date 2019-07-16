package run

import (
	"github.com/pkg/errors"
)

var ErrJobNotFound = errors.New("job not found")

type JobStore struct {
	jobs []Job
}

func NewJobsStore() *JobStore {
	return &JobStore{}
}

func (s *JobStore) Register(j Job) {
	s.jobs = append(s.jobs, j)
}

func (s *JobStore) Jobs() []Job {
	return s.jobs[:]
}

func (s *JobStore) Fetch(n string) (Job, error) {
	for _, j := range s.jobs {
		if j.Name == n {
			return j, nil
		}
	}

	return Job{}, ErrJobNotFound
}
