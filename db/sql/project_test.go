package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectAlertThreadRoundTrip(t *testing.T) {
	store := InitConfigCreateTestStore()

	thread := "42"
	chat := "-1001"
	created, err := store.CreateProject(db.Project{
		Name:        "alerts",
		Alert:       true,
		AlertChat:   &chat,
		AlertThread: &thread,
	})
	require.NoError(t, err)
	require.NotNil(t, created.AlertThread)
	assert.Equal(t, thread, *created.AlertThread)

	got, err := store.GetProject(created.ID)
	require.NoError(t, err)
	require.NotNil(t, got.AlertChat)
	require.NotNil(t, got.AlertThread)
	assert.Equal(t, chat, *got.AlertChat)
	assert.Equal(t, thread, *got.AlertThread)

	updatedThread := "99"
	got.AlertThread = &updatedThread
	require.NoError(t, store.UpdateProject(got))

	got, err = store.GetProject(created.ID)
	require.NoError(t, err)
	require.NotNil(t, got.AlertThread)
	assert.Equal(t, updatedThread, *got.AlertThread)
}
