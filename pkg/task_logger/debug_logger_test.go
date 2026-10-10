package task_logger

import (
	"bytes"
	"os/exec"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureDebugOutput(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	prevOut := log.StandardLogger().Out
	prevLevel := log.GetLevel()
	log.SetOutput(&buf)
	log.SetLevel(log.DebugLevel)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetLevel(prevLevel)
	})

	return &buf
}

func TestDebugLogger_Log(t *testing.T) {
	buf := captureDebugOutput(t)

	logger := DebugLogger{Prefix: "browse"}
	logger.Log("hello")
	logger.Logf("value=%d", 42)
	logger.LogWithTime(time.Now(), "with time")

	out := buf.String()
	assert.Contains(t, out, "level=debug")
	assert.Contains(t, out, "hello")
	assert.Contains(t, out, "value=42")
	assert.Contains(t, out, "with time")
	assert.Contains(t, out, "prefix=browse")
}

func TestDebugLogger_SilentAboveDebugLevel(t *testing.T) {
	buf := captureDebugOutput(t)
	log.SetLevel(log.InfoLevel)

	DebugLogger{}.Log("hidden")

	assert.Empty(t, buf.String())
}

func TestDebugLogger_LogCmd(t *testing.T) {
	buf := captureDebugOutput(t)

	cmd := exec.Command("sh", "-c", "echo out-line; echo err-line 1>&2")
	finish := DebugLogger{}.LogCmd(cmd)

	require.NoError(t, cmd.Run())
	finish()
	finish() // idempotent

	out := buf.String()
	assert.Contains(t, out, "out-line")
	assert.Contains(t, out, "err-line")
}
