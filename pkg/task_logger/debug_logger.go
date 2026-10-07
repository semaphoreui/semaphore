package task_logger

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// DebugLogger is a Logger that writes every message to the process console
// (logrus) at DEBUG level. It is useful for git operations triggered outside
// of a task run (e.g. browsing repository files) when the output should be
// visible in the server log but there is no task log to write to.
//
// Prefix is prepended to every message so that lines from different sources
// can be told apart in the shared server log.
type DebugLogger struct {
	Prefix string
}

func (l DebugLogger) Log(msg string) {
	l.LogWithTime(time.Now(), msg)
}

func (l DebugLogger) Logf(format string, a ...any) {
	l.LogWithTime(time.Now(), fmt.Sprintf(format, a...))
}

func (l DebugLogger) LogWithTime(now time.Time, msg string) {
	entry := log.WithTime(now)
	if l.Prefix != "" {
		entry = entry.WithField("prefix", l.Prefix)
	}
	entry.Debug(msg)
}

func (l DebugLogger) LogfWithTime(now time.Time, format string, a ...any) {
	l.LogWithTime(now, fmt.Sprintf(format, a...))
}

// LogCmd attaches the command's stdout and stderr to the logger and returns
// an idempotent finalizer that must be called after the command stops writing.
func (l DebugLogger) LogCmd(cmd *exec.Cmd) func() {
	stderr, stderrWriter := io.Pipe()
	stdout, stdoutWriter := io.Pipe()
	cmd.Stderr = stderrWriter
	cmd.Stdout = stdoutWriter

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		l.logPipe(stderr)
	}()
	go func() {
		defer wg.Done()
		l.logPipe(stdout)
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			_ = stderrWriter.Close()
			_ = stdoutWriter.Close()
			wg.Wait()
		})
	}
}

func (l DebugLogger) logPipe(reader io.Reader) {
	if closer, ok := reader.(io.Closer); ok {
		defer closer.Close() //nolint:errcheck
	}

	scanner := bufio.NewScanner(reader)
	const maxCapacity = 10 * 1024 * 1024 // 10 MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		l.Log(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		l.Logf("Failed to read command output: %v", err)
	}
}

func (DebugLogger) SetStatus(TaskStatus)             {}
func (DebugLogger) AddStatusListener(StatusListener) {}
func (DebugLogger) AddLogListener(LogListener)       {}
func (DebugLogger) SetCommit(string, string)         {}
