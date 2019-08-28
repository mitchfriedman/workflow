package run_test

import (
	"testing"

	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/stretchr/testify/assert"
)

func TestInputData_GetSliceOfInt64(t *testing.T) {
	tests := map[string]struct {
		have interface{}
		want interface{}
	}{
		"with a list of int64": {[]int64{10}, []int64{10}},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			d := run.InputData{
				"val": tc.have,
			}

			assert.Equal(t, tc.want, d.UnmarshalSliceInt64("val"))
		})
	}
}

func TestInputData_UnmarshalSlice(t *testing.T) {
	tests := map[string]struct {
		have interface{}
		want interface{}
	}{
		"with a list of map of interfaces": {[]map[string]interface{}{{"test": 10}}, []map[string]interface{}{{"test": 10}}},
		"with a single map of interface":   {map[string]interface{}{"test": 10}, []map[string]interface{}{{"test": 10}}},
	}

	convert := func(m map[string]interface{}) interface{} {
		return m
	}

	toSlice := func(d []interface{}) []map[string]interface{} {
		result := make([]map[string]interface{}, len(d), len(d))
		for i, r := range d {
			result[i] = r.(map[string]interface{})
		}

		return result
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			d := run.InputData{
				"val": tc.have,
			}

			result := toSlice(d.UnmarshalSlice("val", convert))
			assert.Equal(t, tc.want, result)
		})
	}
}

func TestInputData_GetString(t *testing.T) {
	tests := map[string]struct {
		have interface{}
		want interface{}
	}{
		"with an int32":  {int32(10), "10"},
		"with an int62":  {int64(10), "10"},
		"with a string":  {"10", "10"},
		"with a float64": {float64(10), "10"},
		"with a float32": {float32(10), "10"},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			d := run.InputData{
				"val": tc.have,
			}

			assert.Equal(t, tc.want, d.UnmarshalString("val"))
		})
	}
}

func TestInputData_GetInt(t *testing.T) {
	tests := map[string]struct {
		have interface{}
		want interface{}
	}{
		"with an int32":  {int32(10), 10},
		"with an int62":  {int64(10), 10},
		"with a string":  {"10", 10},
		"with a float64": {float64(10), 10},
		"with a float32": {float32(10), 10},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			d := run.InputData{
				"val": tc.have,
			}

			assert.Equal(t, tc.want, d.UnmarshalInt("val"))
		})
	}
}
