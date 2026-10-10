package projects

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setScheduleTimezone(t *testing.T, timezone string) {
	t.Helper()

	config := util.Config
	t.Cleanup(func() { util.Config = config })

	util.Config = &util.ConfigType{Schedule: &util.ScheduleConfig{Timezone: timezone}}
}

func validateCronFormatRequest(body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/api/project/1/schedules/validate", strings.NewReader(body))
	w := httptest.NewRecorder()
	ValidateScheduleCronFormat(w, r)
	return w
}

func TestValidateScheduleCronFormat_ReturnsNextRuns(t *testing.T) {
	setScheduleTimezone(t, "UTC")

	w := validateCronFormatRequest(`{"cron_format": "0 3 * * 2#2", "offset_days": 1}`)
	require.Equal(t, http.StatusOK, w.Code)

	var res scheduleNextRuns
	require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
	require.Len(t, res.NextRuns, scheduleNextRunsCount)

	prev := time.Now()
	for _, run := range res.NextRuns {
		assert.True(t, run.After(prev), "runs must be in the future and in order")
		assert.Equal(t, time.Wednesday, run.Weekday())
		assert.Equal(t, 3, run.Hour())
		prev = run
	}
}

func TestValidateScheduleCronFormat_RejectsInvalidSchedules(t *testing.T) {
	setScheduleTimezone(t, "UTC")

	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{
			name:    "both day fields set",
			body:    `{"cron_format": "0 3 1 * 2#2"}`,
			wantErr: "Cron: day of month or day of week must be * or ?",
		},
		{
			name:    "offset out of range",
			body:    `{"cron_format": "0 3 * * 2#2", "offset_days": 99}`,
			wantErr: "Cron: offset must be between -31 and 31 days",
		},
		{
			name:    "offset with @every",
			body:    `{"cron_format": "@every 5h", "offset_days": 1}`,
			wantErr: "Cron: an offset cannot be used with @every",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := validateCronFormatRequest(tt.body)
			require.Equal(t, http.StatusBadRequest, w.Code)

			var res map[string]string
			require.NoError(t, json.NewDecoder(w.Body).Decode(&res))
			assert.Contains(t, res["error"], tt.wantErr)
		})
	}
}

func TestValidateSchedulePayload_OffsetDays(t *testing.T) {
	t.Run("run once schedules drop the cron fields", func(t *testing.T) {
		runAt := time.Now().Add(time.Hour)
		schedule := db.Schedule{
			Type:       db.ScheduleTypeRunAt,
			RunAt:      &runAt,
			CronFormat: "0 3 * * 2#2",
			OffsetDays: 1,
		}

		require.True(t, validateSchedulePayload(&schedule, httptest.NewRecorder()))
		assert.Empty(t, schedule.CronFormat)
		assert.Zero(t, schedule.OffsetDays)
	})

	t.Run("cron schedules keep the offset", func(t *testing.T) {
		schedule := db.Schedule{CronFormat: "0 3 * * 2#2", OffsetDays: -1}

		require.True(t, validateSchedulePayload(&schedule, httptest.NewRecorder()))
		assert.Equal(t, -1, schedule.OffsetDays)
	})

	t.Run("cron schedules reject an offset out of range", func(t *testing.T) {
		schedule := db.Schedule{CronFormat: "0 3 * * 2#2", OffsetDays: 32}
		w := httptest.NewRecorder()

		assert.False(t, validateSchedulePayload(&schedule, w))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
