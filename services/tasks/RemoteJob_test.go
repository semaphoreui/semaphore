package tasks

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestRunnerExclusionReason(t *testing.T) {
	tag := "production"
	now := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)
	recent := now.Add(-time.Minute)
	tests := []struct {
		name         string
		runner       db.Runner
		runnerTag    *string
		runningTasks int
		expected     string
	}{
		{"inactive", db.Runner{}, nil, 0, "inactive"},
		{"not registered", db.Runner{Active: true, IsDefault: true}, nil, 0, "not_registered"},
		{"not default", db.Runner{Active: true, Token: "token", Touched: &recent}, nil, 0, "not_default"},
		{"tag mismatch", db.Runner{Active: true, Token: "token", Tags: []string{"staging"}, Touched: &recent}, &tag, 0, "tag_mismatch"},
		{"offline", db.Runner{Active: true, Token: "token", IsDefault: true}, nil, 0, "offline"},
		{"at capacity", db.Runner{Active: true, Token: "token", IsDefault: true, Touched: &recent, MaxParallelTasks: 2}, nil, 2, "at_capacity"},
		{"available", db.Runner{Active: true, Token: "token", IsDefault: true, Touched: &recent}, nil, 10, ""},
		{"webhook runner", db.Runner{Active: true, Token: "token", IsDefault: true, Webhook: "https://runner.example.test"}, nil, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := runnerExclusionReason(tt.runner, tt.runnerTag, now, 2*time.Minute, tt.runningTasks)
			assert.Equal(t, tt.expected, reason)
		})
	}
}
