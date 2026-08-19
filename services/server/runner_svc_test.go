package server

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerService_CreateRunner_Registered(t *testing.T) {
	svc := NewRunnerService(sql.InitConfigCreateTestStore())

	runner, err := svc.CreateRunner(db.Runner{Name: "r1", Registered: true})
	require.NoError(t, err)

	// A normal runner gets an auth token.
	assert.NotEmpty(t, runner.Token)
	assert.True(t, runner.IsRegistered())
	assert.Nil(t, runner.RegistrationTokenHash)
}

func TestRunnerService_CreateRunner_Unregistered(t *testing.T) {
	svc := NewRunnerService(sql.InitConfigCreateTestStore())

	runner, err := svc.CreateRunner(db.Runner{Registered: false, Active: true})
	require.NoError(t, err)

	// An unregistered runner is created with no credentials at all: no auth token,
	// no registration token, and inactive.
	assert.Empty(t, runner.Token)
	assert.False(t, runner.IsRegistered())
	assert.True(t, runner.Active)
	assert.Nil(t, runner.RegistrationTokenHash)
	assert.Nil(t, runner.RegistrationTokenExpiresAt)
}

func TestRunnerService_RegenerateRegistrationToken(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	svc := NewRunnerService(store)

	runner, err := svc.CreateRunner(db.Runner{Registered: false})
	require.NoError(t, err)
	// Created without any registration token.
	require.Nil(t, runner.RegistrationTokenHash)

	token, err := svc.RegenerateRegistrationToken(runner)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(token, RunnerRegistrationTokenPrefix))

	stored, err := store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	assert.Empty(t, stored.Token)
	require.NotNil(t, stored.RegistrationTokenHash)
	assert.Equal(t, HashRunnerRegistrationToken(token), *stored.RegistrationTokenHash)
	assert.NotNil(t, stored.RegistrationTokenExpiresAt)
}

func TestRunnerService_RegenerateRegistrationToken_ResetsRegisteredRunner(t *testing.T) {
	store := sql.InitConfigCreateTestStore()
	svc := NewRunnerService(store)

	runner, err := svc.CreateRunner(db.Runner{Registered: true, Active: true})
	require.NoError(t, err)
	require.True(t, runner.IsRegistered())

	token, err := svc.RegenerateRegistrationToken(runner)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(token, RunnerRegistrationTokenPrefix))

	stored, err := store.GetGlobalRunner(runner.ID)
	require.NoError(t, err)
	// The runner is reset to the unregistered state.
	assert.Empty(t, stored.Token)
	assert.False(t, stored.IsRegistered())
	//assert.False(t, stored.Active)
	require.NotNil(t, stored.RegistrationTokenHash)
	assert.Equal(t, HashRunnerRegistrationToken(token), *stored.RegistrationTokenHash)
	assert.NotNil(t, stored.RegistrationTokenExpiresAt)
}
