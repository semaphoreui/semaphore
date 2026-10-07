package sql

import (
	"context"
	"fmt"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pageSeqs(page db.AuditEventPage) []int64 {
	var out []int64
	for _, e := range page.Events {
		out = append(out, e.Seq)
	}
	return out
}

func TestListAuditEvents_PagesNewestFirst(t *testing.T) {
	store := InitConfigCreateTestStore()
	for i := 0; i < 7; i++ {
		event := fullAuditEvent()
		event.EventID = fmt.Sprintf("list-%d", i)
		_, err := store.CreateAuditEvent(context.Background(), event)
		require.NoError(t, err)
	}

	latest, err := store.ListAuditEvents(context.Background(), 0, 0, 3)
	require.NoError(t, err)
	assert.Equal(t, []int64{7, 6, 5}, pageSeqs(latest))
	assert.Equal(t, int64(5), latest.Older)
	assert.Zero(t, latest.Newer)

	older, err := store.ListAuditEvents(context.Background(), latest.Older, 0, 3)
	require.NoError(t, err)
	assert.Equal(t, []int64{4, 3, 2}, pageSeqs(older))
	assert.Equal(t, int64(2), older.Older)
	assert.Equal(t, int64(4), older.Newer)

	newer, err := store.ListAuditEvents(context.Background(), 0, older.Newer, 3)
	require.NoError(t, err)
	assert.Equal(t, []int64{7, 6, 5}, pageSeqs(newer))
	assert.Zero(t, newer.Newer)
	assert.Equal(t, int64(5), newer.Older)

	last, err := store.ListAuditEvents(context.Background(), 2, 0, 3)
	require.NoError(t, err)
	assert.Equal(t, []int64{1}, pageSeqs(last))
	assert.Zero(t, last.Older)
}

func TestListAuditEvents_EmptyTable(t *testing.T) {
	store := InitConfigCreateTestStore()

	page, err := store.ListAuditEvents(context.Background(), 0, 0, 50)

	require.NoError(t, err)
	assert.Empty(t, page.Events)
	assert.Zero(t, page.Older)
}
