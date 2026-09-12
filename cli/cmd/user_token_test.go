package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTokenStore implements only the db.Store methods exercised by the token
// commands. The embedded db.Store interface is nil, so any other call panics —
// which is fine because the commands under test never reach them.
type fakeTokenStore struct {
	db.Store

	usersByLogin map[string]db.User
	tokens       map[int][]db.APIToken

	created    []db.APIToken
	getUserErr error
	createErr  error
	listErr    error
}

func (s *fakeTokenStore) GetUserByLoginOrEmail(login string, _ string) (db.User, error) {
	if s.getUserErr != nil {
		return db.User{}, s.getUserErr
	}
	user, ok := s.usersByLogin[login]
	if !ok {
		return db.User{}, db.ErrNotFound
	}
	return user, nil
}

func (s *fakeTokenStore) CreateAPIToken(token db.APIToken) (db.APIToken, error) {
	if s.createErr != nil {
		return db.APIToken{}, s.createErr
	}
	token.Created = tz.Now()
	s.created = append(s.created, token)
	return token, nil
}

func (s *fakeTokenStore) GetAPITokens(userID int) ([]db.APIToken, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.tokens[userID], nil
}

func newFakeTokenStore() *fakeTokenStore {
	return &fakeTokenStore{
		usersByLogin: map[string]db.User{
			"user1": {ID: 42, Username: "user1"},
		},
		tokens: map[int][]db.APIToken{},
	}
}

func TestGetTokenUser(t *testing.T) {
	store := newFakeTokenStore()

	t.Run("missing login", func(t *testing.T) {
		_, err := getTokenUser(store, "")
		assert.ErrorContains(t, err, "--login required")
	})

	t.Run("unknown user", func(t *testing.T) {
		_, err := getTokenUser(store, "ghost")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("found", func(t *testing.T) {
		user, err := getTokenUser(store, "user1")
		require.NoError(t, err)
		assert.Equal(t, 42, user.ID)
	})
}

func TestCreateUserToken_WithTTL(t *testing.T) {
	store := newFakeTokenStore()
	var out bytes.Buffer

	before := tz.Now()
	err := createUserToken(store, &out, tokenArgs{login: "user1", ttl: "1h", name: "ci"})
	require.NoError(t, err)

	require.Len(t, store.created, 1)
	token := store.created[0]

	assert.Equal(t, 42, token.UserID)
	assert.Equal(t, "ci", token.Name)
	assert.False(t, token.Expired)
	require.NotNil(t, token.ExpiresAt)
	assert.True(t, token.ExpiresAt.After(before.Add(59*time.Minute)))
	assert.True(t, token.ExpiresAt.Before(before.Add(61*time.Minute)))

	// create prints the token value
	assert.Equal(t, token.ID, strings.TrimSpace(out.String()))
	assert.NotEmpty(t, token.ID)
}

func TestCreateUserToken_NoTTL_NeverExpires(t *testing.T) {
	store := newFakeTokenStore()
	var out bytes.Buffer

	err := createUserToken(store, &out, tokenArgs{login: "user1"})
	require.NoError(t, err)

	require.Len(t, store.created, 1)
	assert.Nil(t, store.created[0].ExpiresAt)
	assert.Empty(t, store.created[0].Name)
}

func TestCreateUserToken_InvalidTTL(t *testing.T) {
	store := newFakeTokenStore()
	var out bytes.Buffer

	err := createUserToken(store, &out, tokenArgs{login: "user1", ttl: "bogus"})
	assert.ErrorContains(t, err, "invalid --ttl value")
	assert.Empty(t, store.created)
	assert.Empty(t, out.String())
}

func TestCreateUserToken_MissingLogin(t *testing.T) {
	store := newFakeTokenStore()
	var out bytes.Buffer

	err := createUserToken(store, &out, tokenArgs{})
	assert.ErrorContains(t, err, "--login required")
	assert.Empty(t, store.created)
}

func TestCreateUserToken_UnknownUser(t *testing.T) {
	store := newFakeTokenStore()
	var out bytes.Buffer

	err := createUserToken(store, &out, tokenArgs{login: "ghost"})
	assert.ErrorContains(t, err, "not found")
	assert.Empty(t, store.created)
}

func TestListUserTokens_DoesNotLeakTokenValues(t *testing.T) {
	store := newFakeTokenStore()
	expires := tz.Now().Add(time.Hour)
	store.tokens[42] = []db.APIToken{
		{ID: "secret-token-value-active", Name: "ci", ExpiresAt: &expires},
		{ID: "secret-token-value-never", Name: "infra"},
		{ID: "secret-token-value-revoked", Name: "old", Expired: true},
	}

	var out bytes.Buffer
	err := listUserTokens(store, &out, tokenArgs{login: "user1"})
	require.NoError(t, err)

	output := out.String()

	// The token IDs (secret values) must never appear in list output.
	for _, token := range store.tokens[42] {
		assert.NotContains(t, output, token.ID,
			"list output must not leak token value %s", token.ID)
	}

	// Names, statuses and expiry are shown instead.
	assert.Contains(t, output, "ci")
	assert.Contains(t, output, "infra")
	assert.Contains(t, output, "old")
	assert.Contains(t, output, "active")
	assert.Contains(t, output, "expired")
	assert.Contains(t, output, "never")
	assert.Contains(t, output, expires.Format(time.RFC3339))
}

func TestListUserTokens_StatusAndExpiry(t *testing.T) {
	store := newFakeTokenStore()
	expired := tz.Now().Add(-time.Hour)
	active := tz.Now().Add(time.Hour)
	store.tokens[42] = []db.APIToken{
		{ID: "a", Name: "active-ttl", ExpiresAt: &active},
		{ID: "b", Name: "past-ttl", ExpiresAt: &expired},
		{ID: "c", Name: "revoked", Expired: true},
		{ID: "d", Name: "forever"},
	}

	var out bytes.Buffer
	require.NoError(t, listUserTokens(store, &out, tokenArgs{login: "user1"}))

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.Len(t, lines, 4)

	assert.Equal(t, "active-ttl\tactive\t"+active.Format(time.RFC3339), lines[0])
	assert.Equal(t, "past-ttl\texpired\t"+expired.Format(time.RFC3339), lines[1])
	assert.Equal(t, "revoked\texpired\tnever", lines[2])
	assert.Equal(t, "forever\tactive\tnever", lines[3])
}

func TestListUserTokens_Empty(t *testing.T) {
	store := newFakeTokenStore()

	var out bytes.Buffer
	require.NoError(t, listUserTokens(store, &out, tokenArgs{login: "user1"}))
	assert.Empty(t, out.String())
}

func TestListUserTokens_NotFoundTreatedAsEmpty(t *testing.T) {
	store := newFakeTokenStore()
	store.listErr = db.ErrNotFound

	var out bytes.Buffer
	require.NoError(t, listUserTokens(store, &out, tokenArgs{login: "user1"}))
	assert.Empty(t, out.String())
}

func TestListUserTokens_MissingLogin(t *testing.T) {
	store := newFakeTokenStore()

	var out bytes.Buffer
	err := listUserTokens(store, &out, tokenArgs{})
	assert.ErrorContains(t, err, "--login required")
}
