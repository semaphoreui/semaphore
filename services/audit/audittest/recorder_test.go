package audittest

import (
	"context"
	"testing"

	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnly(t *testing.T) {
	rec := &Recorder{}
	ctx := audit.WithRequest(audit.WithActor(context.Background(), audit.SystemActor(audit.ComponentServer)), audit.RequestInfo{ID: "r"})
	rec.Record(ctx, audit.Event{Kind: audit.AuthLogout})

	got, err := rec.Only(audit.AuthLogout)
	require.NoError(t, err)
	assert.Equal(t, audit.SystemActor(audit.ComponentServer), got.Actor)
	assert.Equal(t, "r", got.Request.ID)

	_, err = rec.Only(audit.AuditLifecycleStart)
	assert.Error(t, err, "wrong kind")
}

func TestOnly_ReportsCountAndCatalogViolations(t *testing.T) {
	rec := &Recorder{}
	_, err := rec.Only(audit.AuthLogout)
	assert.Error(t, err, "nothing recorded")

	rec.Record(context.Background(), audit.Event{Kind: audit.AuthLogin})
	_, err = rec.Only(audit.AuthLogin)
	assert.Error(t, err, "missing metadata breaks the catalog")
}

func TestKinds(t *testing.T) {
	rec := &Recorder{}
	kinds, err := rec.Kinds()
	require.NoError(t, err)
	assert.Empty(t, kinds)

	rec.Record(context.Background(), audit.Event{Kind: audit.IAMUserDelete})
	rec.Record(context.Background(), audit.Event{Kind: audit.AuthLogout})
	kinds, err = rec.Kinds()
	require.NoError(t, err)
	assert.Equal(t, []audit.Kind{audit.IAMUserDelete, audit.AuthLogout}, kinds)
}
