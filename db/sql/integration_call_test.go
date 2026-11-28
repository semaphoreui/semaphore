package sql

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestCall(t *testing.T, store *SqlDb, call db.IntegrationCall) {
	t.Helper()
	call.Created = time.Now()
	require.NoError(t, store.CreateIntegrationCall(call))
}

func TestIntegrationCalls_Filters(t *testing.T) {
	store := InitConfigCreateTestStore()

	project, integration := createTestIntegration(t, store)
	other, err := store.CreateProject(db.Project{Name: "other project"})
	require.NoError(t, err)

	createTestCall(t, store, db.IntegrationCall{
		ProjectID: project.ID, IntegrationID: &integration.ID,
		IntegrationName: integration.Name, Alias: "alias-a",
		Method: "POST", Outcome: db.IntegrationCallOK,
	})
	createTestCall(t, store, db.IntegrationCall{
		ProjectID: project.ID, Alias: "alias-b",
		Method: "GET", Outcome: db.IntegrationCallSkipped, Reason: "no match",
	})
	createTestCall(t, store, db.IntegrationCall{
		ProjectID: other.ID, Alias: "alias-a",
		Method: "POST", Outcome: db.IntegrationCallOK,
	})

	t.Run("project scoped, newest first", func(t *testing.T) {
		calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, false)
		require.NoError(t, err)
		require.Len(t, calls, 2)
		assert.Equal(t, "alias-b", calls[0].Alias)
		assert.Equal(t, "alias-a", calls[1].Alias)
	})

	t.Run("by alias", func(t *testing.T) {
		calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "alias-b", nil, false)
		require.NoError(t, err)
		require.Len(t, calls, 1)
		assert.Equal(t, db.IntegrationCallSkipped, calls[0].Outcome)
		assert.Equal(t, "no match", calls[0].Reason)
		assert.Nil(t, calls[0].IntegrationID)
	})

	t.Run("by integration", func(t *testing.T) {
		calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", &integration.ID, false)
		require.NoError(t, err)
		require.Len(t, calls, 1)
		assert.Equal(t, "alias-a", calls[0].Alias)
	})

	t.Run("unknown alias", func(t *testing.T) {
		calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "nope", nil, false)
		require.NoError(t, err)
		assert.Empty(t, calls)
	})
}

func TestIntegrationCalls_Keyset(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestIntegration(t, store)

	for range 3 {
		createTestCall(t, store, db.IntegrationCall{
			ProjectID: project.ID, Alias: "alias", Method: "POST", Outcome: db.IntegrationCallOK,
		})
	}

	first, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{Count: 2}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, first, 2)
	assert.Greater(t, first[0].ID, first[1].ID)

	next, err := store.GetIntegrationCalls(
		project.ID, db.RetrieveQueryParams{Count: 2, BeforeID: first[1].ID}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, next, 1)
	assert.Less(t, next[0].ID, first[1].ID)
}

// A deleted integration must leave its calls readable: that is the whole point
// of storing the alias and the name as plain strings.
func TestIntegrationCalls_SurviveIntegrationDeletion(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, integration := createTestIntegration(t, store)

	createTestCall(t, store, db.IntegrationCall{
		ProjectID: project.ID, IntegrationID: &integration.ID,
		IntegrationName: integration.Name, Alias: "alias", Method: "POST",
		Outcome: db.IntegrationCallOK,
	})

	require.NoError(t, store.DeleteIntegration(project.ID, integration.ID))

	calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, calls, 1)
	assert.Equal(t, "alias", calls[0].Alias)
	assert.Equal(t, integration.Name, calls[0].IntegrationName)
}

func TestIntegrationCall_ReasonTruncated(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestIntegration(t, store)

	long := ""
	for range maxIntegrationCallReason + 50 {
		long += "x"
	}

	createTestCall(t, store, db.IntegrationCall{
		ProjectID: project.ID, Alias: "alias", Method: "POST",
		Outcome: db.IntegrationCallError, Reason: long,
	})

	calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, calls, 1)
	assert.Len(t, calls[0].Reason, maxIntegrationCallReason)
}

func TestTruncateUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected string
	}{
		{"shorter than the limit", "abc", 10, "abc"},
		{"exactly the limit", "abcd", 4, "abcd"},
		{"ascii cut", "abcdef", 3, "abc"},
		{"cut lands inside a rune", "aé", 2, "a"},
		{"cut lands on a rune start", "aéb", 3, "aé"},
		{"every rune is multi-byte", "日本語", 4, "日"},
		{"limit zero", "日本語", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := truncateUTF8(tt.input, tt.n)
			assert.Equal(t, tt.expected, out)
			assert.True(t, utf8.ValidString(out), "truncation must not leave invalid UTF-8")
			assert.LessOrEqual(t, len(out), max(tt.n, 0))
		})
	}
}

// A reason carrying non-ASCII must survive the round trip to the database:
// PostgreSQL rejects an invalid byte sequence outright, which would lose the row.
func TestIntegrationCall_MultiByteReasonIsStored(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestIntegration(t, store)

	long := ""
	for range maxIntegrationCallReason {
		long += "é"
	}

	createTestCall(t, store, db.IntegrationCall{
		ProjectID: project.ID, Alias: "alias", Method: "POST",
		Outcome: db.IntegrationCallError, Reason: long,
	})

	calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, calls, 1)
	assert.True(t, utf8.ValidString(calls[0].Reason))
	assert.LessOrEqual(t, len(calls[0].Reason), maxIntegrationCallReason)
	assert.NotEmpty(t, calls[0].Reason)
}

// The list must stay light: a page of the history would otherwise carry every
// body on it. One call is loaded with its body.
func TestIntegrationCall_BodyOnlyOnTheSingleCall(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestIntegration(t, store)

	body := `{"ref":"refs/heads/main"}`
	createTestCall(t, store, db.IntegrationCall{
		ProjectID: project.ID, Alias: "alias", Method: "POST", Outcome: db.IntegrationCallOK,
		RequestSize: len(body), RequestBody: &body,
	})

	listed, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Nil(t, listed[0].RequestBody, "the list leaves the body out")
	assert.Equal(t, len(body), listed[0].RequestSize, "but keeps its size")

	one, err := store.GetIntegrationCall(project.ID, listed[0].ID)
	require.NoError(t, err)
	require.NotNil(t, one.RequestBody)
	assert.Equal(t, body, *one.RequestBody)
}

// The project is part of the lookup, so a call id from another project — a
// guessable integer — is not found rather than returned.
func TestIntegrationCall_NotReadableFromAnotherProject(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, _ := createTestIntegration(t, store)
	other, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	body := "secret payload"
	createTestCall(t, store, db.IntegrationCall{
		ProjectID: project.ID, Alias: "alias", Method: "POST", Outcome: db.IntegrationCallOK,
		RequestBody: &body,
	})
	calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, calls, 1)

	_, err = store.GetIntegrationCall(other.ID, calls[0].ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func TestIntegrationCall_BodyIsScrubbedAndCapped(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		checkFn func(t *testing.T, stored string)
	}{
		{
			name: "invalid utf-8 is replaced, not rejected",
			body: "ok\xff\xfeend",
			checkFn: func(t *testing.T, stored string) {
				assert.True(t, utf8.ValidString(stored))
				assert.Contains(t, stored, "ok")
				assert.Contains(t, stored, "end")
			},
		},
		{
			name: "NUL bytes are replaced, PostgreSQL rejects them",
			body: "a\x00b",
			checkFn: func(t *testing.T, stored string) {
				assert.Equal(t, "a\uFFFDb", stored)
			},
		},
		{
			name: "a large body is cut at the cap",
			body: strings.Repeat("é", maxIntegrationCallBody),
			checkFn: func(t *testing.T, stored string) {
				assert.LessOrEqual(t, len(stored), maxIntegrationCallBody)
				assert.True(t, utf8.ValidString(stored))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := InitConfigCreateTestStore()
			project, _ := createTestIntegration(t, store)

			body := tt.body
			createTestCall(t, store, db.IntegrationCall{
				ProjectID: project.ID, Alias: "alias", Method: "POST", Outcome: db.IntegrationCallOK,
				RequestSize: len(body), RequestBody: &body,
			})

			calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, true)
			require.NoError(t, err)
			require.Len(t, calls, 1)
			require.NotNil(t, calls[0].RequestBody)
			assert.Equal(t, len(tt.body), calls[0].RequestSize, "the size is the one received")
			tt.checkFn(t, *calls[0].RequestBody)
		})
	}
}

func TestIntegrationCall_UnauthorizedCallsAreCapped(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, integration := createTestIntegration(t, store)

	record := func(outcome db.IntegrationCallOutcome) {
		createTestCall(t, store, db.IntegrationCall{
			ProjectID: project.ID, IntegrationID: &integration.ID,
			IntegrationName: integration.Name, Alias: "alias", Method: "POST", Outcome: outcome,
		})
	}

	record(db.IntegrationCallOK)
	for range maxUnauthorizedIntegrationCalls + 5 {
		record(db.IntegrationCallUnauthorized)
	}

	calls, err := store.GetIntegrationCalls(project.ID, db.RetrieveQueryParams{}, "", nil, false)
	require.NoError(t, err)
	require.Len(t, calls, maxUnauthorizedIntegrationCalls+1, "only rejected calls are capped")

	ok := calls[len(calls)-1]
	assert.Equal(t, db.IntegrationCallOK, ok.Outcome)
	assert.Equal(t, ok.ID+6, calls[len(calls)-2].ID, "the oldest rejected calls go first")
}
