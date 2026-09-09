package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Reproduces "Failed to delete user ... constraint failed: FOREIGN KEY
// constraint failed (787)": on SQLite session.user_id references user(id)
// without an ON DELETE action, so deleting any user who has ever logged in
// (session rows are only flagged expired, never removed) fails.
func TestDeleteUser_WithSession(t *testing.T) {
	store := InitConfigCreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe",
		Name:     "John Doe",
		Email:    "jdoe@example.com",
	})
	require.NoError(t, err)

	now := time.Now()
	_, err = store.CreateSession(db.Session{
		UserID:     user.ID,
		Created:    now,
		LastActive: now,
		IP:         "127.0.0.1",
		UserAgent:  "test",
	})
	require.NoError(t, err)

	require.NoError(t, store.DeleteUser(user.ID))

	_, err = store.GetUser(user.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

// Same failure via task.user_id: the SQLite schema declares
// task.user_id REFERENCES user(id) with no ON DELETE action, so a user who
// has ever run a task cannot be deleted.
func TestDeleteUser_WithTask(t *testing.T) {
	store := InitConfigCreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "jdoe",
		Name:     "John Doe",
		Email:    "jdoe@example.com",
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
		Created:    time.Now(),
	}, 0)
	require.NoError(t, err)

	require.NoError(t, store.DeleteUser(user.ID))

	_, err = store.GetUser(user.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	// The task itself must survive the deletion of its author.
	_, err = store.GetTask(projectID, task.ID)
	assert.NoError(t, err)
}
