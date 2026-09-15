package resp_test

import (
	"bytes"
	"testing"

	"multi-threaded-Redis/Internal/resp"
)

func TestWriter(t *testing.T) {
	tests := []struct {
		name     string
		input    resp.Value
		expected string
	}{
		{
			name: "Simple String",
			input: resp.Value{
				Type: "string",
				Str:  "OK",
			},
			expected: "+OK\r\n",
		},
		{
			name: "Error",
			input: resp.Value{
				Type: "error",
				Str:  "Error message",
			},
			expected: "-Error message\r\n",
		},
		{
			name: "Integer",
			input: resp.Value{
				Type: "integer",
				Num:  1000,
			},
			expected: ":1000\r\n",
		},
		{
			name: "Bulk String",
			input: resp.Value{
				Type: "bulk",
				Bulk: "foobar",
			},
			expected: "$6\r\nfoobar\r\n",
		},
		{
			name: "Array",
			input: resp.Value{
				Type: "array",
				Array: []resp.Value{
					{Type: "bulk", Bulk: "foo"},
					{Type: "bulk", Bulk: "bar"},
				},
			},
			expected: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
		},
		{
			name: "Null",
			input: resp.Value{
				Type: "null",
			},
			expected: "$-1\r\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := resp.NewWriter(&buf)
			err := writer.Write(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if buf.String() != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, buf.String())
			}
		})
	}
}
