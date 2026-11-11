package sql

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (f environmentSecretFixture) secretsOf(t *testing.T, env db.Environment) []db.EnvironmentSecretWithKey {
	secrets, err := f.store.GetEnvironmentSecretKeys(f.project.ID, env.ID)
	require.NoError(t, err)
	return secrets
}

// expireClaim moves the claim of a key into the past, as if the node holding it had died long ago.
func (f environmentSecretFixture) expireClaim(t *testing.T, key db.AccessKey) {
	_, err := f.store.exec("update access_key set delete_claim=? where id=?",
		tz.Now().Add(-time.Minute).UnixMilli(), key.ID)
	require.NoError(t, err)
}

func TestClaimAccessKeyDeletion_RefusedWhileASecretLinksTheKey(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	link := f.link(t, f.otherEnv, db.EnvironmentSecretEnv, "TOKEN", f.shared)

	_, err := f.store.ClaimAccessKeyDeletion(f.project.ID, f.shared.ID, 0)

	assert.ErrorIs(t, err, db.ErrInvalidOperation)
	key, err := f.store.GetAccessKey(f.project.ID, f.shared.ID)
	require.NoError(t, err)
	assert.Nil(t, key.DeleteClaim, "a refused claim leaves nothing behind")

	t.Run("but for the link the key is deleted with", func(t *testing.T) {
		claim, err := f.store.ClaimAccessKeyDeletion(f.project.ID, f.shared.ID, link.ID)
		require.NoError(t, err)
		assert.NotZero(t, claim)
	})
}

func TestClaimAccessKeyDeletion_UnknownKey(t *testing.T) {
	f := newEnvironmentSecretFixture(t)

	_, err := f.store.ClaimAccessKeyDeletion(f.project.ID, 9999, 0)
	assert.ErrorIs(t, err, db.ErrNotFound)

	_, err = f.store.ClaimAccessKeyDeletion(f.project.ID+1, f.shared.ID, 0)
	assert.ErrorIs(t, err, db.ErrNotFound, "a key is claimed only through its own project")
}

// The point of the claim: once it is taken, no secret can be pointed at the key, by a new link or by
// re-pointing an existing one, and nothing is left behind by the refused write.
func TestClaimAccessKeyDeletion_RefusesNewLinksToTheKey(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	existing := f.link(t, f.env, db.EnvironmentSecretVar, "token", f.owned)

	claim, err := f.store.ClaimAccessKeyDeletion(f.project.ID, f.shared.ID, 0)
	require.NoError(t, err)

	_, err = f.store.CreateEnvironmentSecret(db.EnvironmentSecret{
		ProjectID: f.project.ID, EnvironmentID: f.otherEnv.ID, Type: db.EnvironmentSecretEnv, Name: "TOKEN", KeyID: f.shared.ID,
	})
	assert.ErrorIs(t, err, db.ErrInvalidOperation)
	assert.ErrorContains(t, err, "key shared is being deleted")
	assert.Empty(t, f.secretsOf(t, f.otherEnv), "the refused link is not left in the group")

	repointed := existing
	repointed.Name = "renamed"
	repointed.KeyID = f.shared.ID
	err = f.store.UpdateEnvironmentSecret(repointed)
	assert.ErrorIs(t, err, db.ErrInvalidOperation)
	after, err := f.store.GetEnvironmentSecret(f.project.ID, f.env.ID, existing.ID)
	require.NoError(t, err)
	assert.Equal(t, existing, after, "the secret keeps its key and its name")

	// A key nobody deletes is not affected.
	other := f.link(t, f.otherEnv, db.EnvironmentSecretEnv, "OWNED", f.owned)
	assert.NotZero(t, other.ID)

	t.Run("a released claim lets the key be linked again", func(t *testing.T) {
		require.NoError(t, f.store.ReleaseAccessKeyDeletion(f.project.ID, f.shared.ID, claim))
		f.link(t, f.otherEnv, db.EnvironmentSecretEnv, "TOKEN", f.shared)
		require.NoError(t, f.store.UpdateEnvironmentSecret(repointed))
	})
}

// Two deletions of one key: the second is refused rather than allowed to release the claim of the first
// while the first is removing the secret.
func TestClaimAccessKeyDeletion_OneDeletionAtATime(t *testing.T) {
	f := newEnvironmentSecretFixture(t)

	first, err := f.store.ClaimAccessKeyDeletion(f.project.ID, f.shared.ID, 0)
	require.NoError(t, err)

	_, err = f.store.ClaimAccessKeyDeletion(f.project.ID, f.shared.ID, 0)
	assert.ErrorIs(t, err, db.ErrInvalidOperation)
	assert.ErrorContains(t, err, "being deleted")

	t.Run("a claim of somebody else is not released", func(t *testing.T) {
		require.NoError(t, f.store.ReleaseAccessKeyDeletion(f.project.ID, f.shared.ID, first+1))
		key, err := f.store.GetAccessKey(f.project.ID, f.shared.ID)
		require.NoError(t, err)
		require.NotNil(t, key.DeleteClaim)
		assert.Equal(t, first, *key.DeleteClaim)
		assert.True(t, key.DeletionClaimed(tz.Now()))
	})

	t.Run("released, the key can be claimed again", func(t *testing.T) {
		require.NoError(t, f.store.ReleaseAccessKeyDeletion(f.project.ID, f.shared.ID, first))
		_, err := f.store.ClaimAccessKeyDeletion(f.project.ID, f.shared.ID, 0)
		assert.NoError(t, err)
	})
}

// A node that died between the claim and the delete must not lock the key for good.
func TestClaimAccessKeyDeletion_AnExpiredClaimIsNoClaim(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	_, err := f.store.ClaimAccessKeyDeletion(f.project.ID, f.shared.ID, 0)
	require.NoError(t, err)
	f.expireClaim(t, f.shared)

	link := f.link(t, f.env, db.EnvironmentSecretEnv, "TOKEN", f.shared)

	_, err = f.store.ClaimAccessKeyDeletion(f.project.ID, f.shared.ID, link.ID)
	assert.NoError(t, err, "the stale claim is taken over")
}

func TestAccessKey_DeletionClaimed(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) *int64 {
		v := now.Add(d).UnixMilli()
		return &v
	}
	tests := []struct {
		name  string
		claim *int64
		want  bool
	}{
		{"no claim", nil, false},
		{"live claim", at(db.AccessKeyDeleteClaimTTL), true},
		{"claim ending now", at(0), false},
		{"expired claim", at(-time.Second), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := db.AccessKey{DeleteClaim: tt.claim}
			assert.Equal(t, tt.want, key.DeletionClaimed(now))
		})
	}
}

func TestDeleteAccessKeyWithSecret(t *testing.T) {
	t.Run("the key and its link go together", func(t *testing.T) {
		f := newEnvironmentSecretFixture(t)
		link := f.link(t, f.env, db.EnvironmentSecretVar, "token", f.owned)

		require.NoError(t, f.store.DeleteAccessKeyWithSecret(f.project.ID, f.owned.ID, f.env.ID, link.ID))

		assert.Empty(t, f.secretsOf(t, f.env))
		_, err := f.store.GetAccessKey(f.project.ID, f.owned.ID)
		assert.ErrorIs(t, err, db.ErrNotFound)
	})

	t.Run("a second link keeps both", func(t *testing.T) {
		f := newEnvironmentSecretFixture(t)
		link := f.link(t, f.env, db.EnvironmentSecretVar, "token", f.owned)
		f.link(t, f.otherEnv, db.EnvironmentSecretVar, "borrowed", f.owned)

		err := f.store.DeleteAccessKeyWithSecret(f.project.ID, f.owned.ID, f.env.ID, link.ID)

		assert.ErrorIs(t, err, db.ErrInvalidOperation)
		assert.Len(t, f.secretsOf(t, f.env), 1, "the link is rolled back with the key")
		_, err = f.store.GetAccessKey(f.project.ID, f.owned.ID)
		assert.NoError(t, err)
	})

	t.Run("a link to another key is not this key's link", func(t *testing.T) {
		f := newEnvironmentSecretFixture(t)
		link := f.link(t, f.env, db.EnvironmentSecretVar, "token", f.shared)

		err := f.store.DeleteAccessKeyWithSecret(f.project.ID, f.owned.ID, f.env.ID, link.ID)

		assert.ErrorIs(t, err, db.ErrNotFound)
		assert.Len(t, f.secretsOf(t, f.env), 1)
		_, err = f.store.GetAccessKey(f.project.ID, f.owned.ID)
		assert.NoError(t, err)
	})
}

// The link write reads the key back through its project, so a key of another project is refused by the
// store too, not only by the service in front of it.
func TestEnvironmentSecret_KeyOfAnotherProjectIsRefused(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	otherProject, err := f.store.CreateProject(db.Project{Name: "Q"})
	require.NoError(t, err)
	secret := "x"
	foreign, err := f.store.CreateAccessKey(db.AccessKey{
		Name: "foreign", Type: db.AccessKeyString, ProjectID: &otherProject.ID, Secret: &secret,
	})
	require.NoError(t, err)

	_, err = f.store.CreateEnvironmentSecret(db.EnvironmentSecret{
		ProjectID: f.project.ID, EnvironmentID: f.env.ID, Type: db.EnvironmentSecretVar, Name: "x", KeyID: foreign.ID,
	})

	var validation *common_errors.ValidationError
	require.ErrorAs(t, err, &validation)
	assert.ErrorContains(t, err, "does not exist in this project")
	assert.Empty(t, f.secretsOf(t, f.env))
}
