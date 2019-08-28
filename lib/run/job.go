package run

import (
	"fmt"
	"strconv"
)

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

func (d InputData) UnmarshalSliceInt64(field string) []int64 {
	val, ok := d[field]
	if !ok {
		return []int64{}
	}

	var vals []int64
	for _, d := range val.([]int64) {
		vals = append(vals, d)
	}

	return vals
}

type converter func(m map[string]interface{}) interface{}

func (d InputData) UnmarshalSlice(field string, c converter) []interface{} {
	var result []interface{}

	val, ok := d[field]
	if !ok {
		return result
	}

	asMap, ok := val.(map[string]interface{})
	if ok {
		result = append(result, c(asMap))
		return result
	}

	for _, d := range val.([]map[string]interface{}) {
		result = append(result, c(d))
	}

	return result
}

func (d InputData) UnmarshalString(field string) string {
	val := d[field]
	switch val.(type) {
	case int32:
		return strconv.Itoa(int(val.(int32)))
	case int64:
		return strconv.Itoa(int(val.(int64))) // can lose precision here
	case string:
		return val.(string)
	case float64:
		return strconv.Itoa(int(val.(float64)))
	case float32:
		return strconv.Itoa(int(val.(float32)))
	default:
		return fmt.Sprintf("%s", val)
	}
}

func (d InputData) UnmarshalInt(field string) int {
	val := d[field]
	switch val.(type) {
	case int:
		return val.(int)
	case int32:
		return int(val.(int32))
	case int64:
		return int(val.(int64))
	case float64:
		return int(val.(float64)) // can lose precision here.
	case float32:
		return int(val.(float32))
	case string:
		v, err := strconv.Atoi(val.(string))
		if err != nil {
			return 0
		}
		return v
	default:
		return 0
	}
}

const InputTypeString = "string"
const InputTypeInt = "int"
const InputTypeList = "list"

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
