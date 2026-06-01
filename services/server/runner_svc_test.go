package server

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerService_CreateRunner_Registered(t *testing.T) {
	svc := NewRunnerService(bolt.CreateTestStore())

	runner, privateKey, err := svc.CreateRunner(db.Runner{Name: "r1", Registered: true})
	require.NoError(t, err)

	// A normal runner gets an auth token and a server-generated key pair.
	assert.NotEmpty(t, runner.Token)
	assert.True(t, runner.IsRegistered())
	assert.NotEmpty(t, privateKey)
	assert.NotNil(t, runner.PublicKey)
	assert.Empty(t, runner.RegistrationToken)
	assert.Nil(t, runner.RegistrationTokenHash)
}

func TestRunnerService_CreateRunner_RegisteredWithProvidedPublicKey(t *testing.T) {
	svc := NewRunnerService(bolt.CreateTestStore())

	pub := "provided-public-key"
	runner, privateKey, err := svc.CreateRunner(db.Runner{PublicKey: &pub, Registered: true})
	require.NoError(t, err)

	// When the caller provides a public key, the service does not generate one.
	assert.NotEmpty(t, runner.Token)
	assert.Empty(t, privateKey)
	require.NotNil(t, runner.PublicKey)
	assert.Equal(t, pub, *runner.PublicKey)
}

func TestRunnerService_CreateRunner_Unregistered(t *testing.T) {
	svc := NewRunnerService(bolt.CreateTestStore())

	runner, privateKey, err := svc.CreateRunner(db.Runner{Registered: false, Active: true})
	require.NoError(t, err)

	// An unregistered runner has no auth token, is forced inactive, gets no key
	// pair, and receives a one-time registration token instead.
	assert.Empty(t, runner.Token)
	assert.False(t, runner.IsRegistered())
	assert.False(t, runner.Active)
	assert.Empty(t, privateKey)
	assert.Nil(t, runner.PublicKey)
	assert.NotEmpty(t, runner.RegistrationToken)
	assert.True(t, strings.HasPrefix(runner.RegistrationToken, RunnerRegistrationTokenPrefix))
	require.NotNil(t, runner.RegistrationTokenHash)
	assert.Equal(t, HashRunnerRegistrationToken(runner.RegistrationToken), *runner.RegistrationTokenHash)
	assert.NotNil(t, runner.RegistrationTokenExpiresAt)
}
