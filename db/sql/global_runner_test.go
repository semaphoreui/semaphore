package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterRunner_RefusalErrors(t *testing.T) {
	store := InitConfigCreateTestStore()
	past := tz.Now().Add(-time.Hour)
	future := tz.Now().Add(time.Hour)
	expiredHash, registeredHash := "expired-hash", "registered-hash"
	_, err := store.CreateRunner(db.Runner{Name: "expired", RegistrationTokenHash: &expiredHash, RegistrationTokenExpiresAt: &past})
	require.NoError(t, err)
	_, err = store.CreateRunner(db.Runner{Name: "registered", Token: "runner-token", RegistrationTokenHash: &registeredHash, RegistrationTokenExpiresAt: &future})
	require.NoError(t, err)

	_, err = store.RegisterRunner("unknown-hash", nil)
	assert.ErrorIs(t, err, db.ErrNotFound)

	_, err = store.RegisterRunner(expiredHash, nil)
	assert.ErrorIs(t, err, db.ErrRegistrationTokenExpired)
	assert.EqualError(t, err, "registration token expired")

	_, err = store.RegisterRunner(registeredHash, nil)
	assert.ErrorIs(t, err, db.ErrRunnerAlreadyRegistered)
	assert.EqualError(t, err, "runner is already registered")
}
