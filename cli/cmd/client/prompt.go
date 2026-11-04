package client

import (
	"bufio"
	"context"
	"errors"
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
func (p *Prompter) Confirm(ctx context.Context, question string) error {
	if p.yes {
		return nil
	}
	if p.noInput || !p.interactive {
		return usage(p.cmdPath, "%s: confirmation required; pass --yes to proceed without a prompt", question)
	}
	_, _ = fmt.Fprintf(p.out, "%s [y/N]: ", question)
	answer, err := p.readLine(ctx)
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
func (p *Prompter) ConfirmName(ctx context.Context, what, expected, typed string) error {
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
	answer, err := p.readLine(ctx)
	if err != nil {
		return err
	}
	if answer != expected {
		return ErrAborted
	}
	return nil
}

// canAsk says whether a question can be put to a human right now. --yes is
// deliberately not consulted: it waives the client's own confirmations, and
// approving a task that is about to change infrastructure is the server's
// gate, not one of them.
func (p *Prompter) canAsk() bool {
	return p.interactive && !p.noInput
}

// promptAnswer is one line typed in answer to a question asked while
// something else is going on.
type promptAnswer struct {
	text string
	// eof is set when the input ended before an answer was given.
	eof bool
	err error
}

// readAsync reads one line in the background, so that a caller can keep
// working — polling a task, printing its log — while nobody has answered yet.
// It writes nothing itself: the question and every reply belong to the caller's
// goroutine, which keeps the output of a concurrent reader free of races.
func (p *Prompter) readAsync() <-chan promptAnswer {
	ch := make(chan promptAnswer, 1)
	go func() {
		line, err := p.in.ReadString('\n')
		line = strings.TrimSpace(line)
		switch {
		case err == nil || (errors.Is(err, io.EOF) && line != ""):
			ch <- promptAnswer{text: line}
		case errors.Is(err, io.EOF):
			ch <- promptAnswer{eof: true}
		default:
			ch <- promptAnswer{err: err}
		}
	}()
	return ch
}

// readLine waits for one line, or for ctx to be done. Reading stdin directly
// would ignore the cancellation: the client traps SIGINT to end a wait
// cleanly, so a blocking read here would swallow Ctrl+C and leave the command
// hanging on a question.
func (p *Prompter) readLine(ctx context.Context) (string, error) {
	select {
	case answer := <-p.readAsync():
		if answer.err != nil {
			return "", answer.err
		}
		return answer.text, nil
	case <-ctx.Done():
		// The question stays on the line the user was typing on.
		_, _ = fmt.Fprintln(p.out)
		return "", ctx.Err()
	}
}
