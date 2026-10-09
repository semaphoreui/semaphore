package sql

import (
	"slices"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

func TestMigration_2_20_11_DropIndexes(t *testing.T) {
	tests := []struct {
		dialect string
		format  string
	}{
		{"mysql", "drop index %s on audit_event"},
		{util.DbDriverSQLite, "drop index if exists %s"},
		{"postgres", "drop index if exists %s"},
	}
	names := []string{"audit_event_actor_idx", "audit_event_project_idx", "audit_event_kind_idx",
		"audit_event_outcome_idx", "audit_event_ip_idx", "audit_event_target_idx"}
	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			sql := strings.Join(getVersionSQL(tt.dialect, "v2.20.11.err.sql", false), "\n")
			for _, name := range names {
				assert.Contains(t, sql, strings.Replace(tt.format, "%s", name, 1))
			}
		})
	}
}

func TestMigration_2_20_11_CreatedIndexHasSeq(t *testing.T) {
	tests := []struct {
		dialect string
		drop    string
	}{
		{"mysql", "drop index audit_event_created_idx on audit_event"},
		{util.DbDriverSQLite, "drop index if exists audit_event_created_idx"},
		{"postgres", "drop index if exists audit_event_created_idx"},
	}
	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			for file, create := range map[string]string{
				"v2.20.11.sql":     "create index audit_event_created_idx on audit_event (created, seq)",
				"v2.20.11.err.sql": "create index audit_event_created_idx on audit_event (created)",
			} {
				queries := getVersionSQL(tt.dialect, file, false)
				drop := slices.Index(queries, tt.drop)
				assert.GreaterOrEqual(t, drop, 0, file)
				assert.Greater(t, slices.Index(queries, create), drop, file)
			}
		})
	}
}
