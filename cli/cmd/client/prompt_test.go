package client

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPrompter(input string, interactive, yes, noInput bool) (*Prompter, *bytes.Buffer) {
	var out bytes.Buffer
	return &Prompter{
		in:          bufio.NewReader(strings.NewReader(input)),
		out:         &out,
		interactive: interactive,
		yes:         yes,
		noInput:     noInput,
		cmdPath:     "semaphore client template delete",
	}, &out
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		interactive bool
		yes         bool
		noInput     bool
		wantErr     error
		wantUsage   bool
		wantPrompt  bool
	}{
		{"--yes skips the prompt even without a TTY", "", false, true, false, nil, false, false},
		{"not a TTY: usage error", "y\n", false, false, false, nil, true, false},
		{"--no-input in a TTY: usage error", "y\n", true, false, true, nil, true, false},
		{"TTY answers yes", "y\n", true, false, false, nil, false, true},
		{"TTY answers YES", "YES\n", true, false, false, nil, false, true},
		{"TTY answers no", "n\n", true, false, false, ErrAborted, false, true},
		{"TTY empty answer is no", "\n", true, false, false, ErrAborted, false, true},
		{"TTY EOF is no", "", true, false, false, ErrAborted, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, out := newPrompter(tt.input, tt.interactive, tt.yes, tt.noInput)
			err := p.Confirm("Delete template 5?")
			if tt.wantUsage {
				var u *UsageError
				require.ErrorAs(t, err, &u)
				assert.Contains(t, err.Error(), "--yes")
				assert.Equal(t, ExitUsage, ExitCode(err))
				return
			}
			assert.Equal(t, tt.wantErr, err)
			if tt.wantPrompt {
				assert.Equal(t, "Delete template 5? [y/N]: ", out.String())
			} else {
				assert.Empty(t, out.String())
			}
		})
	}
}

func TestConfirmName(t *testing.T) {
	p, _ := newPrompter("", false, true, false)
	assert.NoError(t, p.ConfirmName("project", "prod", "prod"), "--confirm matching wins everywhere")

	err := p.ConfirmName("project", "prod", "prodd")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	assert.Contains(t, err.Error(), "does not match")

	p, _ = newPrompter("", false, true, false)
	err = p.ConfirmName("project", "prod", "")
	require.ErrorAs(t, err, &u, "--yes is not enough for a project deletion")
	assert.Contains(t, err.Error(), `--confirm="prod"`)

	p, out := newPrompter("prod\n", true, false, false)
	assert.NoError(t, p.ConfirmName("project", "prod", ""))
	assert.Contains(t, out.String(), `Type the project name "prod" to confirm: `)

	p, _ = newPrompter("other\n", true, false, false)
	assert.Equal(t, ErrAborted, p.ConfirmName("project", "prod", ""))
}
