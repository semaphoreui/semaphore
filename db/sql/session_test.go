package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteAPIToken_ReturnsOnlyTheTokensItDeleted(t *testing.T) {
	store := InitConfigCreateTestStore()
	user, err := store.CreateUser(db.UserWithPwd{Pwd: "verystrongpassword1", User: db.User{Username: "alice", Name: "alice", Email: "alice@example.com"}})
	require.NoError(t, err)
	for _, id := range []string{"abcdefgh1111", "abcdefgh2222", "zzzzzzzz3333"} {
		_, err = store.CreateAPIToken(db.APIToken{ID: id, UserID: user.ID, Name: id})
		require.NoError(t, err)
	}

	deleted, err := store.DeleteAPIToken(user.ID, "abcdefgh")
	require.NoError(t, err)
	var ids []string
	for _, token := range deleted {
		ids = append(ids, token.ID)
	}
	assert.ElementsMatch(t, []string{"abcdefgh1111", "abcdefgh2222"}, ids)

	deleted, err = store.DeleteAPIToken(user.ID, "abcdefgh")
	require.NoError(t, err)
	assert.Empty(t, deleted, "a token deleted earlier is not reported again")

	remaining, err := store.GetAPITokens(user.ID)
	require.NoError(t, err)
	require.Len(t, remaining, 1)
	assert.Equal(t, "zzzzzzzz3333", remaining[0].ID)
}
