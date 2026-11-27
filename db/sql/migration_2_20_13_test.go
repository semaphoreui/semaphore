package sql

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

// The workflow survey columns are added with plain `alter table ... add
// column`, which every dialect accepts; the test pins the three columns and
// their rollback so a renamed column can not drift from db.Workflow* tags.
func TestMigration_2_20_13_WorkflowSurveyColumns(t *testing.T) {
	dialects := []string{"mysql", util.DbDriverSQLite, "postgres"}

	up := []string{
		"project__workflow_revision",
		"survey_vars",
		"project__workflow_node",
		"survey_inputs",
		"project__workflow_run",
		"environment",
	}

	for _, dialect := range dialects {
		t.Run(dialect, func(t *testing.T) {
			queries := getVersionSQL(dialect, "v2.20.13.sql", false)
			joined := strings.Join(queries, "\n")
			for _, want := range up {
				assert.Contains(t, joined, want)
			}
			assert.Equal(t, 3, strings.Count(strings.ToLower(joined), "add column"))

			rollback := strings.Join(getVersionSQL(dialect, "v2.20.13.err.sql", false), "\n")
			assert.Equal(t, 3, strings.Count(strings.ToLower(rollback), "drop column"))
			for _, want := range up {
				assert.Contains(t, rollback, want)
			}
		})
	}
}
