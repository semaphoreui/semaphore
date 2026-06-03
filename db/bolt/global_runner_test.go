package bolt

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/services/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetRunnerByToken_ReturnsGlobalRunnerWhenTokenExists(t *testing.T) {
	store := CreateTestStore()

	testRunner, err := store.CreateRunner(db.Runner{Token: db.GenerateRunnerToken()})
	require.NoError(t, err)

	_, err = store.GetRunnerByToken(testRunner.Token)
	assert.NoError(t, err)
}

func Test_GetRunnerByToken_ReturnsRunnerWhenTokenExists(t *testing.T) {
	store := CreateTestStore()

	project, err := store.CreateProject(db.Project{})
	require.NoError(t, err)

	testRunner, err := store.CreateRunner(db.Runner{ProjectID: &project.ID, Token: db.GenerateRunnerToken()})
	require.NoError(t, err)

	_, err = store.GetRunnerByToken(testRunner.Token)
	assert.NoError(t, err)
}

func Test_GetGlobalRunner_ReturnsErrorWhenTryingGetProjectRunner(t *testing.T) {
	store := CreateTestStore()

	project, err := store.CreateProject(db.Project{})
	assert.NoError(t, err)

	testRunner, err := store.CreateRunner(db.Runner{ProjectID: &project.ID})
	assert.NoError(t, err)

	_, err = store.GetGlobalRunner(testRunner.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

// CreateRunner is a pure persister: it stores the registration-token hash and
// expiry it is given without altering them. Credential generation itself lives
// in the RunnerService.
func Test_CreateRunner_PersistsRegistrationTokenFields(t *testing.T) {
	store := CreateTestStore()

	hash := server.HashRunnerRegistrationToken("plaintext-token")
	expiresAt := tz.Now().Add(time.Hour)

	testRunner, err := store.CreateRunner(db.Runner{
		RegistrationTokenHash:      &hash,
		RegistrationTokenExpiresAt: &expiresAt,
	})
	require.NoError(t, err)
	assert.Empty(t, testRunner.Token)

	stored, err := store.GetGlobalRunner(testRunner.ID)
	require.NoError(t, err)
	assert.Empty(t, stored.Token)
	require.NotNil(t, stored.RegistrationTokenHash)
	assert.Equal(t, hash, *stored.RegistrationTokenHash)
	assert.NotNil(t, stored.RegistrationTokenExpiresAt)
}

func Test_RegisterRunner_SetsTokenAndActivates(t *testing.T) {
	store := CreateTestStore()

	hash := server.HashRunnerRegistrationToken("sms_test")
	expiresAt := tz.Now().Add(time.Hour)
	testRunner, err := store.CreateRunner(db.Runner{
		RegistrationTokenHash:      &hash,
		RegistrationTokenExpiresAt: &expiresAt,
	})
	require.NoError(t, err)

	publicKey := "test-public-key"
	registered, err := store.RegisterRunner(hash, &publicKey, true)
	require.NoError(t, err)
	assert.NotEmpty(t, registered.Token)
	assert.True(t, registered.Active)
	assert.Equal(t, publicKey, *registered.PublicKey)

	stored, err := store.GetGlobalRunner(testRunner.ID)
	require.NoError(t, err)
	assert.Equal(t, registered.Token, stored.Token)
	assert.True(t, stored.Active)
	// The one-time registration token is cleared after use.
	assert.Nil(t, stored.RegistrationTokenHash)
	assert.Nil(t, stored.RegistrationTokenExpiresAt)
}

func Test_RegisterRunner_FailsWhenTokenUnknown(t *testing.T) {
	store := CreateTestStore()

	_, err := store.RegisterRunner(server.HashRunnerRegistrationToken("sms_nope"), nil, true)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func Test_RegisterRunner_FailsWhenExpired(t *testing.T) {
	store := CreateTestStore()

	hash := server.HashRunnerRegistrationToken("sms_expired")
	expiresAt := tz.Now().Add(-time.Hour)
	_, err := store.CreateRunner(db.Runner{
		RegistrationTokenHash:      &hash,
		RegistrationTokenExpiresAt: &expiresAt,
	})
	require.NoError(t, err)

	_, err = store.RegisterRunner(hash, nil, true)
	assert.Error(t, err)
}

func Test_ResetRunnerRegistration_DeactivatesRunner(t *testing.T) {
	store := CreateTestStore()

	runner, err := store.CreateRunner(db.Runner{
		Token:  db.GenerateRunnerToken(),
		Active: true,
	})
	require.NoError(t, err)
	require.True(t, runner.Active)

	hash := server.HashRunnerRegistrationToken("sms_reset")
	expiresAt := tz.Now().Add(time.Hour)
	err = store.ResetRunnerRegistration(runner.ID, hash, expiresAt)
	require.NoError(t, err)

	stored, err := store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	assert.Empty(t, stored.Token)
	assert.False(t, stored.Active)
	require.NotNil(t, stored.RegistrationTokenHash)
	assert.Equal(t, hash, *stored.RegistrationTokenHash)
}
