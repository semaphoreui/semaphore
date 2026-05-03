package bolt

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression test for semaphoreui/semaphore#3580.
//
// When a project was restored more than once (or whenever an integration alias
// happened to collide with one already stored in the global public alias bucket),
// publicAlias.createAlias passed an uninitialised `any` to getPublicAlias. That
// nil destination propagated into unmarshalObject -> createObjectType, where
// reflect.TypeOf(nil) yielded a nil reflect.Type and the call panicked with a
// nil pointer dereference. The expected behaviour is a normal duplicate-alias
// error, not a server-killing panic.
func TestCreateIntegrationAlias_DuplicateDoesNotPanic(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)

	_, err = store.CreateIntegrationAlias(db.IntegrationAlias{
		Alias:     "shared-alias",
		ProjectID: proj.ID,
	})
	require.NoError(t, err)

	require.NotPanics(t, func() {
		_, err = store.CreateIntegrationAlias(db.IntegrationAlias{
			Alias:     "shared-alias",
			ProjectID: proj.ID,
		})
	})
	assert.Error(t, err, "duplicate alias must surface as an error")
}

// TestUnmarshalObject_NilDestinationReturnsError guards the defensive check
// added to unmarshalObject so that any future caller that forgets to allocate
// a destination object gets a clean error instead of a panic.
func TestUnmarshalObject_NilDestinationReturnsError(t *testing.T) {
	require.NotPanics(t, func() {
		err := unmarshalObject([]byte(`{}`), nil, nil)
		assert.Error(t, err)
	})
}
