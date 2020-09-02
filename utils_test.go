package percook

import (
	"fmt"
	"testing"
)

func TestStringCoalesceWithDefault(t *testing.T) {
	for _, tc := range []struct {
		def string
		in  []string
		out string
	}{
		{
			"default",
			[]string{},
			"default",
		},
		{
			"default",
			[]string{"0"},
			"0",
		},
		{
			"default",
			[]string{"", "1"},
			"1",
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("(%s, %v) => %s", tc.def, tc.in, tc.out), func(t *testing.T) {
			actual := stringCoalesceWithDefault(tc.def, tc.in...)
			if tc.out != actual {
				t.Errorf(`"%s" expected but "%s" returned`, tc.out, actual)
			}
		})
	}
}

func TestStringMin(t *testing.T) {
	for _, tc := range []struct {
		in  []string
		out string
	}{
		{
			[]string{},
			"",
		},
		{
			[]string{"0"},
			"0",
		},
		{
			[]string{"", "1"},
			"",
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%v => %s", tc.in, tc.out), func(t *testing.T) {
			actual := stringMin(tc.in...)
			if tc.out != actual {
				t.Errorf(`"%s" expected but "%s" returned`, tc.out, actual)
			}
		})
	}
}
