package runners

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"github.com/semaphoreui/semaphore/services/runners"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingTouchRepo struct {
	db.RunnerManager
	touches int
}

func (r *countingTouchRepo) TouchRunner(runner db.Runner) error {
	r.touches++
	return nil
}

func pollRunner(t *testing.T, runner db.Runner, startedAt *time.Time) (touches int, state runners.RunnerState) {
	t.Helper()

	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })

	store := sql.InitConfigCreateTestStore()
	pool := tasks.CreateTaskPool(store, tasks.NewMemoryTaskStateStore(), nil, nil, nil, nil, nil, nil, nil)
	repo := &countingTouchRepo{RunnerManager: store}
	ctrl := NewRunnerController(repo, &pool, nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/api/internal/runners", nil)
	if startedAt != nil {
		r.Header.Set("X-Runner-Started-At", startedAt.Format(time.RFC3339))
	}
	r = helpers.SetContextValue(r, "runner", runner)
	w := httptest.NewRecorder()

	ctrl.GetRunner(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &state))
	return repo.touches, state
}

func TestGetRunner_SkipsTouchWhileHeartbeatIsFresh(t *testing.T) {
	touched := tz.Now().Add(-time.Second)
	startedAt := tz.Now().Add(-time.Hour).Truncate(time.Second)

	touches, _ := pollRunner(t, db.Runner{ID: 1, Touched: &touched, StartedAt: &startedAt}, &startedAt)

	assert.Equal(t, 0, touches)
}

func TestGetRunner_TouchesWhenHeartbeatIsStale(t *testing.T) {
	touched := tz.Now().Add(-maxRunnerTouchInterval)
	startedAt := tz.Now().Add(-time.Hour).Truncate(time.Second)

	touches, _ := pollRunner(t, db.Runner{ID: 1, Touched: &touched, StartedAt: &startedAt}, &startedAt)

	assert.Equal(t, 1, touches)
}

func TestGetRunner_TouchesNeverTouchedRunner(t *testing.T) {
	touches, _ := pollRunner(t, db.Runner{ID: 1}, nil)

	assert.Equal(t, 1, touches)
}

func TestGetRunner_TouchesWhenRunnerRestarted(t *testing.T) {
	touched := tz.Now().Add(-time.Second)
	startedAt := tz.Now().Add(-time.Hour).Truncate(time.Second)
	restartedAt := tz.Now().Truncate(time.Second)

	touches, _ := pollRunner(t, db.Runner{ID: 1, Touched: &touched, StartedAt: &startedAt}, &restartedAt)

	assert.Equal(t, 1, touches)
}

func TestGetRunner_TouchesWhenCacheCleaningIsDelivered(t *testing.T) {
	touched := tz.Now().Add(-2 * time.Second)
	cleaningRequested := tz.Now().Add(-time.Second)
	startedAt := tz.Now().Add(-time.Hour).Truncate(time.Second)

	touches, state := pollRunner(t, db.Runner{
		ID:                1,
		Touched:           &touched,
		StartedAt:         &startedAt,
		CleaningRequested: &cleaningRequested,
	}, &startedAt)

	assert.True(t, state.ClearCache)
	assert.Equal(t, 1, touches, "touched must pass cleaning_requested so the request is delivered once")
}
