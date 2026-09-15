package resp_test

import (
	"bytes"
	"reflect"
	"testing"

	"multi-threaded-Redis/Internal/resp"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected resp.Value
	}{
		{
			name:  "Simple String",
			input: "+OK\r\n",
			expected: resp.Value{
				Type: "string",
				Str:  "OK",
			},
		},
		{
			name:  "Error",
			input: "-Error message\r\n",
			expected: resp.Value{
				Type: "error",
				Str:  "Error message",
			},
		},
		{
			name:  "Integer",
			input: ":1000\r\n",
			expected: resp.Value{
				Type: "integer",
				Num:  1000,
			},
		},
		{
			name:  "Bulk String",
			input: "$6\r\nfoobar\r\n",
			expected: resp.Value{
				Type: "bulk",
				Bulk: "foobar",
			},
		},
		{
			name:  "Array",
			input: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			expected: resp.Value{
				Type: "array",
				Array: []resp.Value{
					{Type: "bulk", Bulk: "foo"},
					{Type: "bulk", Bulk: "bar"},
				},
			},
		},
		{
			name:  "Null Bulk String",
			input: "$-1\r\n",
			expected: resp.Value{
				Type: "null",
			},
		},
		{
			name:  "Null Array",
			input: "*-1\r\n",
			expected: resp.Value{
				Type: "null",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			parser := resp.NewParser(buf)
			val, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(val, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, val)
			}
		})
	}
}
