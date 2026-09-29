package main

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	page := render()

	assert.True(t, strings.HasPrefix(page, "---\ntitle: Audit events\n"))
	assert.Contains(t, page, "Generated from the Semaphore source")
	for _, entry := range audit.Catalog() {
		assert.Contains(t, page, "| `"+entry.Kind.Code()+"` | `"+entry.Kind.Action()+"` |")
	}
	assert.Contains(t, page, "`method`, `provider`", "metadata fields come from JSON tags")
	assert.Contains(t, page, "| `denied` | failure | `token_unknown`, `token_expired` |", "a denial is only a failure")
	assert.Contains(t, page, "| `change` | success, failure | `invalid_current_password` |", "allowed reasons per event")
	assert.Contains(t, page, "| `activation_failed` |  | Pro |", "Pro-only events are marked")
	assert.Contains(t, page, "| `iam.user` | `create` | `creation` | success |  | `admin`, `pro`, `external` |  |", "OSS events have no edition")
}
