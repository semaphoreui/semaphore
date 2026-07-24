package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEvent_StoresAuditFields(t *testing.T) {
	store := CreateTestStore()

	user, err := store.CreateUserWithoutPassword(db.User{
		Username: "auditor", Name: "Auditor", Email: "auditor@example.com",
	})
	require.NoError(t, err)

	objectType := db.EventUser
	action := "login_success"
	ip := "203.0.113.7"
	userAgent := "Mozilla/5.0"
	description := "User auditor logged in"
	integrationID := 42

	_, err = store.CreateEvent(db.Event{
		UserID:        &user.ID,
		IntegrationID: &integrationID,
		ObjectType:    &objectType,
		ObjectID:      &user.ID,
		Description:   &description,
		Action:        &action,
		IP:            &ip,
		UserAgent:     &userAgent,
	})
	require.NoError(t, err)

	events, err := store.GetAllEvents(db.RetrieveQueryParams{})
	require.NoError(t, err)
	require.Len(t, events, 1)

	evt := events[0]
	require.NotNil(t, evt.Action)
	assert.Equal(t, action, *evt.Action)
	require.NotNil(t, evt.IP)
	assert.Equal(t, ip, *evt.IP)
	require.NotNil(t, evt.UserAgent)
	assert.Equal(t, userAgent, *evt.UserAgent)
	require.NotNil(t, evt.IntegrationID)
	assert.Equal(t, integrationID, *evt.IntegrationID)
}
