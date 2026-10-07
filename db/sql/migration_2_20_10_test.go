package sql

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

func TestMigration_2_20_10_DropIndex(t *testing.T) {
	tests := []struct {
		dialect string
		want    string
	}{
		{"mysql", "drop index audit_event_created_idx on audit_event"},
		{util.DbDriverSQLite, "drop index if exists audit_event_created_idx"},
		{"postgres", "drop index if exists audit_event_created_idx"},
	}
	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			queries := getVersionSQL(tt.dialect, "v2.20.10.err.sql", false)
			assert.Contains(t, strings.Join(queries, "\n"), tt.want)
		})
	}
}
