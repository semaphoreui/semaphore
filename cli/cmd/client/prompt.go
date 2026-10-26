package client

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Prompter asks the user to confirm a destructive action. It never asks when
// the session is not interactive: a script that forgot --yes gets a usage error
// instead of a hang on a question nobody will see.
type Prompter struct {
	in          *bufio.Reader
	out         io.Writer // stderr: a prompt is not data
	interactive bool      // stdin and stdout are terminals
	yes         bool      // --yes given
	noInput     bool      // --no-input, SEMAPHORE_CLIENT_NO_INPUT or CI
	cmdPath     string
}

// Confirm asks "<question> [y/N]". It returns nil when the action may proceed,
// ErrAborted when the user declined, and a UsageError when no question can be
// asked.
func (p *Prompter) Confirm(question string) error {
	if p.yes {
		return nil
	}
	if p.noInput || !p.interactive {
		return usage(p.cmdPath, "%s: confirmation required; pass --yes to proceed without a prompt", question)
	}
	_, _ = fmt.Fprintf(p.out, "%s [y/N]: ", question)
	answer, err := p.readLine()
	if err != nil {
		return err
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return nil
	}
	return ErrAborted
}

// ConfirmName asks the user to type a name back, for the heaviest deletions.
// typed is the value of --confirm when given; it skips the prompt when it
// matches and fails when it does not.
func (p *Prompter) ConfirmName(what, expected, typed string) error {
	if typed != "" {
		if typed == expected {
			return nil
		}
		return usage(p.cmdPath, "--confirm=%q does not match the %s name %q", typed, what, expected)
	}
	if p.noInput || !p.interactive {
		return usage(p.cmdPath, "deleting %s %q requires --confirm=%q", what, expected, expected)
	}
	_, _ = fmt.Fprintf(p.out, "Type the %s name %q to confirm: ", what, expected)
	answer, err := p.readLine()
	if err != nil {
		return err
	}
	if answer != expected {
		return ErrAborted
	}
	return nil
}

func (p *Prompter) readLine() (string, error) {
	line, err := p.in.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
