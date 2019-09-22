package tools

import (
	"errors"
	"reflect"
	"testing"
)

func TestConcurrentAction(t *testing.T) {
	var cases = []struct {
		intention     string
		maxConcurrent uint
		log           bool
		action        func(interface{}) (interface{}, error)
		inputs        []interface{}
		want          []ConcurentOutput
	}{
		{
			"simple input",
			0,
			false,
			func(i interface{}) (interface{}, error) {
				return i, nil
			},
			[]interface{}{
				8000,
			},
			[]ConcurentOutput{
				{Input: 8000, Output: 8000},
			},
		},
		{
			"multiple input with error and high concurrent",
			5,
			true,
			func(i interface{}) (interface{}, error) {
				if i.(int) != 8000 {
					return nil, errors.New("invalid value")
				}

				return i, nil
			},
			[]interface{}{
				8000,
				1000,
			},
			[]ConcurentOutput{
				{Input: 8000, Output: 8000},
				{Input: 1000, Output: nil, Err: errors.New("invalid value")},
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			inputs, results := ConcurrentAction(testCase.maxConcurrent, testCase.log, testCase.action)

			for _, input := range testCase.inputs {
				inputs <- input
			}
			close(inputs)

			outputs := make([]ConcurentOutput, 0)
			for {
				output, ok := <-results
				if !ok {
					break
				}

				outputs = append(outputs, output)
			}

			for _, output := range outputs {
				found := false
				for _, wantOutput := range testCase.want {
					if reflect.DeepEqual(output, wantOutput) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("ConcurrentAction() = (%#v), not found", output)
				}
			}
		})
	}
}
