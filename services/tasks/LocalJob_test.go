package tasks

import (
	"testing"
)

func TestIsCLIArgsOverridden(t *testing.T) {
	var args []struct {
		tmplArg string
		taskArg string
		want    bool
		err     error
	} = []struct {
		tmplArg string
		taskArg string
		want    bool
		err     error
	}{
		{
			tmplArg: "--ssh-extra-args=\"-p 3222\"",
			taskArg: "--ssh-extra-args \"-p 3222\"",
			want:    false,
			err:     nil,
		},
		{
			tmplArg: "--ssh-extra-args=\"-p 3222\"",
			taskArg: "--ssh-extra-args \"-p 3223\"",
			want:    true,
			err:     nil,
		},
		{
			tmplArg: "--ssh-extra-args=\"-p 3222\"",
			taskArg: "--ssh-extra-args=\"-p 3223\"",
			want:    true,
			err:     nil,
		},
		{
			tmplArg: "--ssh-extra-args=\"-p 3222\"",
			taskArg: "--ssh-extra-args",
			want:    false,
			err:     errCliOverrideParseError,
		},
	}

	for _, tc := range args {

		got, err := isCLIArgsOverridden(tc.tmplArg, tc.taskArg)
		if err != tc.err {
			t.Fail()
		}

		if got != tc.want {
			t.Fail()
		}
	}
}
