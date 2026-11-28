package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The activity of a project reaches a member only with read_project; events without a project reach everyone.
func TestGetUserEvents_NeedsReadOnTheProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)
	member := func(name string, mask db.ProjectUserPermission) db.User {
		role, err := store.CreateRole(db.Role{Slug: name, Name: name, Permissions: mask, ProjectID: &project.ID})
		require.NoError(t, err)
		user, err := store.CreateUserWithoutPassword(db.User{Username: name, Name: name, Email: name + "@example.com"})
		require.NoError(t, err)
		require.NoError(t, store.CreateProjectUser(project.ID, user.ID, []int{role.ID}))
		return user
	}
	reader, nobody := member("reader", db.CanReadProject), member("nobody", 0)
	inProject, global := "in project", "global"
	_, err = store.CreateEvent(db.Event{ProjectID: &project.ID, Description: &inProject})
	require.NoError(t, err)
	_, err = store.CreateEvent(db.Event{Description: &global})
	require.NoError(t, err)

	descriptions := func(user db.User) []string {
		events, err := store.GetUserEvents(user.ID, db.RetrieveQueryParams{})
		require.NoError(t, err)
		got := []string{}
		for _, event := range events {
			got = append(got, *event.Description)
		}
		return got
	}
	assert.Equal(t, []string{"global", "in project"}, descriptions(reader))
	assert.Equal(t, []string{"global"}, descriptions(nobody))
}
