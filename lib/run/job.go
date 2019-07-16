package run

type InputData map[string]interface{}

// Merge immutably merges 2 InputData maps by creating a new one and
// copying the original data into it. Then, it copies the merging data in,
// and returns a new InputData type, leaving each InputData unchanged.
func (d InputData) Merge(other InputData) InputData {
	n := make(InputData)
	for k, v := range d {
		n[k] = v
	}

	for k, v := range other {
		n[k] = v
	}

	return n
}

const InputTypeString = "string"
const InputTypeInt = "int"

// Job is a definition of pipeline of work to perform.
type Job struct {
	Name  string `json:"name"`
	Start *Step  `json:"start"`
}

func NewJob(name string, start *Step) Job {
	return Job{
		Name:  name,
		Start: start,
	}
}
