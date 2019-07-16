package engine

import (
	"fmt"
	"sort"

	"github.com/mitchfriedman/workflow/lib/run"

	"github.com/pkg/errors"
)

// Prioritize receives a run Retriever and will determine the best run to execute based on currently executing runs,
// in progress runs, and not-yet-started runs.
func Prioritize(retriever run.Retriever) (*run.Run, error) {
	runs, err := retriever.NextRuns()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch next runs")
	}

	runQueues := make(map[string][]*run.Run)

	keyName := func(r *run.Run) string {
		return fmt.Sprintf("%s-%s", r.JobName, r.Scope)
	}

	// there is a queue of runs by the combination of the job name and the
	// scope of the run. In other words, different scopes of the same job
	// name can run concurrently.
	for _, r := range runs {
		k := keyName(r)
		runQueues[k] = append(runQueues[k], r)
	}

	// TODO: this is depending on map iteration order to ensure that
	// there is no starvation in different job types. If there is starvation,
	// this can be a culprit. Perhaps refactor to use better randomization.
	// The idea here is that we make a list of each run based on their unique run type.
	for _, rs := range runQueues {
		// The idea is that we sort the runs based on their job+scope. If there is a run that we can execute in the that
		// sorted list, it will be the first one in the list and we will choose it. Otherwise, try a different list.
		// The sort order is as follows:
		// 1. Currently claimed runs
		// 2. Already started runs.
		// 3. Earliest to be created.
		sort.Slice(rs, func(i, j int) bool {
			if rs[i].ClaimedBy != nil {
				return true
			} else if rs[j].ClaimedBy != nil {
				return false
			}

			if rs[i].LastStepComplete != nil {
				return true
			} else if rs[j].LastStepComplete != nil {
				return false
			}

			return rs[i].Started.Before(rs[j].Started)
		})

		// make sure that other runs of the same job + scope are queued behind the currently executing one so that
		// only one run of the job+scope is being executed at time.
		if len(rs) > 0 && rs[0].ClaimedBy == nil {
			return rs[0], nil
		}
	}

	// no runs to execute - not an error.
	return nil, nil
}
