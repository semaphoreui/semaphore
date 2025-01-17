package bolt

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/require"
)

func TestBoltDb_UpdateProjectUser(t *testing.T) {
	store := CreateTestStore()

	usr, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "denguk@example.com",
			Name:     "Denis Gukov",
			Username: "fiftin",
		},
	})
	require.NoError(t, err)

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})
	require.NoError(t, err)

	projUser, err := store.CreateProjectUser(db.ProjectUser{
		ProjectID: proj1.ID,
		UserID:    usr.ID,
		Role:      db.ProjectOwner,
	})
	require.NoError(t, err)

	projUser.Role = db.ProjectOwner
	err = store.UpdateProjectUser(projUser)
	require.NoError(t, err)
}

func TestGetUsers(t *testing.T) {
	store := CreateTestStore()

	_, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "denguk@example.com",
			Name:     "Denis Gukov",
			Username: "fiftin",
		},
	})
	require.NoError(t, err)

	found, err := store.GetUsers(db.RetrieveQueryParams{})
	require.NoError(t, err)

	require.Equal(t, 1, len(found))
}

func TestGetUser(t *testing.T) {
	store := CreateTestStore()

	usr, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "denguk@example.com",
			Name:     "Denis Gukov",
			Username: "fiftin",
		},
	})
	require.NoError(t, err)

	found, err := store.GetUser(usr.ID)
	require.NoError(t, err)

	require.Equal(t, "fiftin", found.Username)

	err = store.DeleteUser(usr.ID)
	require.NoError(t, err)
}

func TestGetUserCount(t *testing.T) {
	store := CreateTestStore()

	// Create first user
	_, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "user1@example.com",
			Name:     "User One",
			Username: "userone",
		},
	})
	require.NoError(t, err)

	// Create second user
	_, err = store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "user2@example.com",
			Name:     "User Two",
			Username: "usertwo",
		},
	})
	require.NoError(t, err)

	// Get user count
	count, err := store.GetUserCount()
	require.NoError(t, err)

	// Verify the count
	require.Equal(t, 2, count)
}

func TestBoltDb_DeleteUser(t *testing.T) {
	store := CreateTestStore()

	// Create a user
	usr, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "deleteuser@example.com",
			Name:     "Delete User",
			Username: "deleteuser",
		},
	})
	require.NoError(t, err)

	// Create a project
	proj, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "DeleteUserProject",
	})
	require.NoError(t, err)

	// Associate the user with the project
	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: proj.ID,
		UserID:    usr.ID,
		Role:      db.ProjectOwner,
	})
	require.NoError(t, err)

	// Delete the user
	err = store.DeleteUser(usr.ID)
	require.NoError(t, err)

	// Verify the user is deleted
	_, err = store.GetUser(usr.ID)
	require.Error(t, err)

	// Verify the project-user association is deleted
	_, err = store.GetProjectUser(proj.ID, usr.ID)
	require.Error(t, err)
}
