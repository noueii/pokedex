package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  Hello world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "PIKACHU    CHARIZARD",
			expected: []string{"pikachu", "charizard"},
		},
		{
			input:    "   multiple   spaces   between   ",
			expected: []string{"multiple", "spaces", "between"},
		},
		{
			input:    "tabs\tand\tspaces",
			expected: []string{"tabs", "and", "spaces"},
		},
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "   ",
			expected: []string{},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]

			if word != expectedWord {
				t.Errorf("FAIL: expected - '%s' actual '%s'", expectedWord, word)
				return

			}
		}

	}
}
