package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_2_20_2_EnforcesUniqueTemplateName checks the index exists after
// the migrations the test store runs, so that a name identifies one template.
func TestMigration_2_20_2_EnforcesUniqueTemplateName(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	insert := func(name string) error {
		_, err := store.exec(
			"insert into project__template (project_id, repository_id, name, playbook, arguments, allow_override_args_in_task, app) values (?, ?, ?, 'site.yml', '[]', false, 'ansible')",
			projectID, repositoryID, name)
		return err
	}

	require.NoError(t, insert("Build website"))

	// The store also rejects this, but here the insert goes straight to the
	// database, so only the index can stop it.
	assert.Error(t, insert("Build website"))

	assert.NoError(t, insert("Deploy website"))
}

// TestMigration_2_20_2_RenamesExistingDuplicates covers PreApply: templates which
// already share a name must be renamed rather than failing the upgrade.
func TestMigration_2_20_2_RenamesExistingDuplicates(t *testing.T) {
	store := InitConfigCreateTestStore()
	projectID, repositoryID := newTemplateTestProject(t, store)

	// Recreate the pre-migration state: drop the index, then write the duplicates
	// an old installation would be holding.
	_, err := store.exec("drop index project__template__project_id_name")
	require.NoError(t, err)

	names := []string{"Build", "Build", "Build", "Build (2)", "Deploy"}
	for _, name := range names {
		_, err = store.exec(
			"insert into project__template (project_id, repository_id, name, playbook, arguments, allow_override_args_in_task, app) values (?, ?, ?, 'site.yml', '[]', false, 'ansible')",
			projectID, repositoryID, name)
		require.NoError(t, err)
	}

	tx, err := store.Sql().Begin()
	require.NoError(t, err)

	require.NoError(t, migration_2_20_2{db: store}.PreApply(tx))
	require.NoError(t, tx.Commit())

	var renamed []struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	_, err = store.Sql().Select(&renamed,
		"select id, name from project__template order by id")
	require.NoError(t, err)

	// The oldest template of each name keeps it; the rest get a suffix, skipping
	// "Build (2)" because that name is already taken.
	actual := make([]string, 0, len(renamed))
	for _, template := range renamed {
		actual = append(actual, template.Name)
	}
	assert.Equal(t, []string{"Build", "Build (3)", "Build (4)", "Build (2)", "Deploy"}, actual)

	// The renames must leave the table indexable.
	_, err = store.exec(
		"create unique index project__template__project_id_name on project__template (project_id, name)")
	assert.NoError(t, err)
}
