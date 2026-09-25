package db_lib

import (
	"fmt"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

const (
	// gitStderrLimit caps how much of a git command's stderr is kept for its
	// error. git explains a failure at the end, so the tail is what is kept.
	gitStderrLimit = 16 * 1024

	// gitErrorDetailLines is how many of the last stderr lines an error keeps.
	// It is enough for a "remote:" explanation from the server followed by
	// git's own "fatal:" line.
	gitErrorDetailLines = 5
)

// GitCommandError is returned when a git command fails. Detail holds the last
// lines git wrote to stderr with the repository's credentials redacted, so the
// error is safe to log and to show to the user.
type GitCommandError struct {
	// Command is the git subcommand, such as "ls-remote" or "clone".
	Command string
	Detail  string
	Err     error
}

func (e *GitCommandError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("git %s: %v", e.Command, e.Err)
	}
	return fmt.Sprintf("git %s: %s", e.Command, e.Detail)
}

func (e *GitCommandError) Unwrap() error {
	return e.Err
}

func newGitCommandError(repo db.Repository, args []string, stderr string, err error) *GitCommandError {
	command := ""
	if len(args) > 0 {
		command = args[0]
	}

	return &GitCommandError{
		Command: command,
		Detail:  gitErrorDetail(repo.RedactCredentials(stderr)),
		Err:     err,
	}
}

// gitErrorDetail returns the last non-empty lines of stderr. Carriage returns
// count as line breaks, since that is how git redraws its progress lines.
// git clone's "Cloning into '<dir>'..." is dropped: it explains nothing and
// names a server-side directory.
func gitErrorDetail(stderr string) string {
	stderr = strings.ToValidUTF8(stderr, "")
	stderr = strings.ReplaceAll(stderr, "\r", "\n")

	var lines []string
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Cloning into ") {
			continue
		}
		lines = append(lines, line)
	}

	if len(lines) > gitErrorDetailLines {
		lines = lines[len(lines)-gitErrorDetailLines:]
	}

	return strings.Join(lines, "\n")
}

// tailBuffer is an io.Writer which keeps the last limit bytes written to it.
type tailBuffer struct {
	limit     int
	buf       []byte
	truncated bool
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)

	if over := len(b.buf) - b.limit; over > 0 {
		b.buf = append(b.buf[:0], b.buf[over:]...)
		b.truncated = true
	}

	return len(p), nil
}

// String returns what was kept. When earlier output was dropped, the first,
// partial line is dropped too: it could hold the end of a credential whose
// start was cut off, which RedactCredentials would no longer recognize.
func (b *tailBuffer) String() string {
	if b == nil {
		return ""
	}

	s := string(b.buf)

	if b.truncated {
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = s[i+1:]
		} else {
			s = ""
		}
	}

	return s
}
