package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The view setting is added on every dialect. The two-column index is for
// PostgreSQL only: MySQL and SQLite keep the primary key in every index, so
// task (template_id) already serves there.
func TestMigration_2_21_6_Statements(t *testing.T) {
	const addColumn = "alter table `project__view` add `show_template_stats` boolean not null default false"
	const dropColumn = "alter table `project__view` drop column `show_template_stats`"

	statements := func(dialect string, file string) []string {
		res := []string{}
		for _, query := range getVersionSQL(dialect, file, false) {
			if query != "" {
				res = append(res, query)
			}
		}
		return res
	}

	tests := []struct {
		name     string
		dialect  string
		file     string
		expected []string
	}{
		{"postgres apply", util.DbDriverPostgres, "v2.21.6.sql", []string{
			addColumn,
			"create index if not exists task__template_id__id on task (template_id, id)",
			"drop index if exists task_template_id_idx",
		}},
		{"postgres rollback", util.DbDriverPostgres, "v2.21.6.err.sql", []string{
			dropColumn,
			"create index if not exists task_template_id_idx on task (template_id)",
			"drop index if exists task__template_id__id",
		}},
		{"mysql apply", util.DbDriverMySQL, "v2.21.6.sql", []string{addColumn}},
		{"mysql rollback", util.DbDriverMySQL, "v2.21.6.err.sql", []string{dropColumn}},
		{"sqlite apply", util.DbDriverSQLite, "v2.21.6.sql", []string{addColumn}},
		{"sqlite rollback", util.DbDriverSQLite, "v2.21.6.err.sql", []string{dropColumn}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, statements(tt.dialect, tt.file))
		})
	}
}

// A view that existed before the migration does not show the counts.
func TestMigration_2_21_6_ViewsDoNotShowCountsByDefault(t *testing.T) {
	before := "2.20.20"
	store := InitConfigCreateTestStoreAt(&before)

	project, err := store.CreateProject(db.Project{Name: "old views"})
	require.NoError(t, err)
	_, err = store.exec("insert into project__view (project_id, title, position, `type`, hidden) values (?, 'Old', 0, '', false)", project.ID)
	require.NoError(t, err)

	target := "2.21.6"
	require.NoError(t, db.Migrate(store, &target))

	views, err := store.GetViews(project.ID, nil)
	require.NoError(t, err)
	require.Len(t, views, 1)
	assert.False(t, views[0].ShowTemplateStats)
}
