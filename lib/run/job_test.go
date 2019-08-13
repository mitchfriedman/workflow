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

			assert.Equal(t, tc.want, d.GetSliceOfInt64("val"))
		})
	}
}

func TestInputData_GetSliceOfMaps(t *testing.T) {
	tests := map[string]struct {
		have interface{}
		want interface{}
	}{
		"with a list of map of interfaces": {[]map[string]interface{}{{"test": 10}}, []map[string]interface{}{{"test": 10}}},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			d := run.InputData{
				"val": tc.have,
			}

			assert.Equal(t, tc.want, d.GetSliceOfMaps("val"))
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

			assert.Equal(t, tc.want, d.GetString("val"))
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

			assert.Equal(t, tc.want, d.GetInt("val"))
		})
	}
}
