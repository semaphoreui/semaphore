package debuglog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_Active(t *testing.T) {
	tests := []struct {
		name   string
		spec   string
		active bool
	}{
		{"empty", "", false},
		{"whitespace only", "   \t", false},
		{"single include", "runner", true},
		{"only exclude", "-db", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.active, Parse(tt.spec).Active())
		})
	}
}

func TestFilter_Enabled(t *testing.T) {
	tests := []struct {
		name      string
		spec      string
		namespace string
		enabled   bool
	}{
		{"exact match", "runner", "runner", true},
		{"exact no match", "runner", "task_pool", false},
		{"empty spec disables", "", "runner", false},
		{"multiple includes first", "runner,task_pool", "runner", true},
		{"multiple includes second", "runner,task_pool", "task_pool", true},
		{"multiple includes miss", "runner,task_pool", "git", false},
		{"space separated", "runner task_pool", "task_pool", true},
		{"prefix wildcard match", "task_*", "task_pool", true},
		{"prefix wildcard match logger", "task_*", "task_logger", true},
		{"prefix wildcard miss", "task_*", "runner", false},
		{"global wildcard", "*", "anything", true},
		{"global wildcard empty ns", "*", "", true},
		{"narrow filter empty ns", "runner", "", false},
		{"prefix wildcard empty ns", "task_*", "", false},
		{"exclude wins", "*,-task_pool", "task_pool", false},
		{"exclude lets others through", "*,-task_pool", "runner", true},
		{"exclude wildcard", "*,-task_*", "task_logger", false},
		{"exclude wildcard others", "*,-task_*", "runner", true},
		{"only exclude no include", "-db", "runner", false},
		{"middle wildcard", "a*z", "abcz", true},
		{"middle wildcard miss", "a*z", "abc", false},
		{"dot is literal", "task.pool", "taskXpool", false},
		{"dot literal match", "task.pool", "task.pool", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.enabled, Parse(tt.spec).Enabled(tt.namespace))
		})
	}
}
