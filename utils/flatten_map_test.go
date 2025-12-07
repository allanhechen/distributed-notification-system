package utils

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFlattenMap(t *testing.T) {
	tests := []struct {
		input  map[string]any
		output []any
	}{
		{input: map[string]any{"value1": "hello"}, output: []any{"value1", "hello"}},
		{input: map[string]any{"value2": 1}, output: []any{"value2", 1}},
		{input: map[string]any{"value1": "hello", "value2": 1}, output: []any{"value1", "hello", "value2", 1}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.input), func(t *testing.T) {
			output := FlattenMap(tt.input)
			if diff := cmp.Diff(tt.output, output); diff != "" {
				t.Errorf("got %v, wanted %v", output, tt.output)
			}
		})
	}
}
