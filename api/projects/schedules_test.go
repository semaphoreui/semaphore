package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupScheduleConfig points util.Config.Schedule at a fixed timezone for the
// test and restores the previous value afterwards.
func setupScheduleConfig(t *testing.T) {
	t.Helper()
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
	orig := util.Config.Schedule
	t.Cleanup(func() { util.Config.Schedule = orig })
	util.Config.Schedule = &util.ScheduleConfig{Timezone: "UTC"}
}

// postValidate posts body to the validate handler and returns the recorder.
func postValidate(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/project/1/schedules/validate", strings.NewReader(body))
	w := httptest.NewRecorder()
	ValidateScheduleCronFormat(w, req)
	return w
}

func TestValidateScheduleCronFormat(t *testing.T) {
	setupScheduleConfig(t)

	t.Run("monthly-weekday descriptor returns next runs", func(t *testing.T) {
		w := postValidate(t, `{"cron_format":"@monthly-weekday 2 tue offset 1 at 03:00"}`)
		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			NextRun []time.Time `json:"next_run"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.NextRun, scheduleValidateNextRuns)
		assert.True(t, resp.NextRun[0].After(time.Now()))
		assert.True(t, resp.NextRun[1].After(resp.NextRun[0]))
	})

	t.Run("standard cron returns next runs", func(t *testing.T) {
		w := postValidate(t, `{"cron_format":"* * 1 * *"}`)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "next_run")
	})

	t.Run("invalid cron is rejected", func(t *testing.T) {
		w := postValidate(t, `{"cron_format":"* * * *"}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Cron:")
	})

	t.Run("invalid descriptor is rejected", func(t *testing.T) {
		w := postValidate(t, `{"cron_format":"@monthly-weekday 9 wed"}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "ordinal")
	})
}
