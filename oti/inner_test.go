package oti

import (
	"errors"
	"testing"
)

func TestFirstErrPart(t *testing.T) {
	cases := []struct {
		name string
		in   error
		exp  string
	}{
		{"nil error returns empty", nil, ""},
		{"simple error no delimiter", errors.New("simple error"), "simple error"},
		{"colon delimiter", errors.New("failed to parse: invalid syntax"), "failed to parse"},
		{"semicolon delimiter", errors.New("connect; timeout exceeded"), "connect"},
		{"comma delimiter", errors.New("multi, part error"), "multi"},
		{"multiple delimiters first wins", errors.New("already contains something: detail: inner"), "already contains something"},
		{"starts with delimiter yields empty", errors.New(":oops"), ""},
		{"space before delimiter preserved", errors.New("foo bar ; baz"), "foo bar "},
	}

	for _, tc := range cases {
		got := FirstErrPart(tc.in)
		if got != tc.exp {
			// Show full error string for easier debugging
			var inStr string
			if tc.in != nil {
				inStr = tc.in.Error()
			}

			// Use t.Errorf not fatal to see all failures at once
			t.Errorf("%s: expected '%s' got '%s' (input=%q)", tc.name, tc.exp, got, inStr)
		}
	}
}
