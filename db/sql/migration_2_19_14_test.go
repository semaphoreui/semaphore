package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigration_2_19_14_DataSurvivesRebuild seeds data on the 2.20.1 schema
// and then applies 2.19.14 to ensure the session/task rebuild keeps existing
// rows intact — in particular dropping the old task table must not trigger
// ON DELETE CASCADE into task__output.
func TestMigration_2_19_14_DataSurvivesRebuild(t *testing.T) {
	preRebuild := "2.19.12"
	store := InitConfigCreateTestStoreAt(&preRebuild)

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe", Name: "John Doe", Email: "jdoe@example.com",
	})
	require.NoError(t, err)

	now := time.Now()
	session, err := store.CreateSession(db.Session{
		UserID:     user.ID,
		Created:    now,
		LastActive: now,
		UserAgent:  "test",
	})
	require.NoError(t, err)

	projectID, repositoryID := newTemplateTestProject(t, store)
	template, err := store.CreateTemplate(db.Template{
		ProjectID:    projectID,
		RepositoryID: repositoryID,
		Name:         "tpl",
		Playbook:     "site.yml",
	})
	require.NoError(t, err)

	task, err := store.CreateTask(db.Task{
		TemplateID: template.ID,
		ProjectID:  projectID,
		Status:     "success",
		Playbook:   "site.yml",
		UserID:     &user.ID,
		Created:    now,
	}, 0)
	require.NoError(t, err)

	_, err = store.CreateTaskOutput(db.TaskOutput{
		TaskID: task.ID,
		Time:   now,
		Output: "ok",
	})
	require.NoError(t, err)

	// Apply the remaining migrations (2.19.14 rebuilds session and task).
	require.NoError(t, db.Migrate(store, nil))

	// The rebuild runs with foreign_keys=OFF, so make sure it left no
	// dangling references behind.
	violations, err := store.Sql().SelectInt("select count(1) from pragma_foreign_key_check")
	require.NoError(t, err)
	assert.Zero(t, violations)

	// Existing rows survived the rebuild.
	survivedSession, err := store.GetSession(user.ID, session.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, survivedSession.UserID)

	survivedTask, err := store.GetTask(projectID, task.ID)
	require.NoError(t, err)
	require.NotNil(t, survivedTask.UserID)
	assert.Equal(t, user.ID, *survivedTask.UserID)

	outputs, err := store.GetTaskOutputs(projectID, task.ID, db.RetrieveQueryParams{})
	require.NoError(t, err)
	assert.Len(t, outputs, 1)

	// And the rebuilt FKs are in effect: a bare user delete cascades the
	// session and nulls task.user_id.
	_, err = store.exec("delete from `user` where id=?", user.ID)
	require.NoError(t, err)

	_, err = store.GetSession(user.ID, session.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	survivedTask, err = store.GetTask(projectID, task.ID)
	require.NoError(t, err)
	assert.Nil(t, survivedTask.UserID)
}
