package schedules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustLoadLocation(t testing.TB, name string) *time.Location {
	t.Helper()

	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}

func nextRunsRFC3339(t *testing.T, cronFormat string, loc *time.Location, from time.Time, n int) []string {
	t.Helper()

	schedule, err := ParseCronSchedule(cronFormat, loc)
	require.NoError(t, err)

	runs := make([]string, 0, n)
	for _, run := range NextRuns(schedule, from, n) {
		runs = append(runs, run.Format(time.RFC3339))
	}
	return runs
}

func TestParseCronSchedule_NextRuns(t *testing.T) {
	from := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		cronFormat string
		want       []string
	}{
		{
			name:       "standard expression",
			cronFormat: "*/60 * * * *",
			want:       []string{"2026-10-08T13:00:00Z", "2026-10-08T14:00:00Z"},
		},
		{
			name:       "day of month or day of week, as before",
			cronFormat: "0 0 1 * 1",
			want:       []string{"2026-10-12T00:00:00Z", "2026-10-19T00:00:00Z", "2026-10-26T00:00:00Z", "2026-11-01T00:00:00Z"},
		},
		{
			name:       "second Tuesday",
			cronFormat: "0 3 * * 2#2",
			want:       []string{"2026-10-13T03:00:00Z", "2026-11-10T03:00:00Z", "2026-12-08T03:00:00Z"},
		},
		{
			name:       "weekday name and question mark",
			cronFormat: "0 3 ? * TUE#2",
			want:       []string{"2026-10-13T03:00:00Z", "2026-11-10T03:00:00Z", "2026-12-08T03:00:00Z"},
		},
		{
			name:       "first and third Monday",
			cronFormat: "0 3 * * 1#1,1#3",
			want:       []string{"2026-10-19T03:00:00Z", "2026-11-02T03:00:00Z", "2026-11-16T03:00:00Z"},
		},
		{
			name:       "last Friday",
			cronFormat: "30 4 * * 5L",
			want:       []string{"2026-10-30T04:30:00Z", "2026-11-27T04:30:00Z", "2026-12-25T04:30:00Z"},
		},
		{
			name:       "last day of the month",
			cronFormat: "0 3 L * *",
			want:       []string{"2026-10-31T03:00:00Z", "2026-11-30T03:00:00Z", "2026-12-31T03:00:00Z"},
		},
		{
			name:       "last weekday of the month",
			cronFormat: "0 3 LW * *",
			want:       []string{"2026-10-30T03:00:00Z", "2026-11-30T03:00:00Z", "2026-12-31T03:00:00Z"},
		},
		{
			name:       "weekday nearest to the 15th",
			cronFormat: "0 3 15W * *",
			want:       []string{"2026-10-15T03:00:00Z", "2026-11-16T03:00:00Z", "2026-12-15T03:00:00Z"},
		},
		{
			name:       "time fields read as robfig/cron reads them",
			cronFormat: "*/20 3-4 * NOV-DEC 2#2",
			want:       []string{"2026-11-10T03:00:00Z", "2026-11-10T03:20:00Z", "2026-11-10T03:40:00Z", "2026-11-10T04:00:00Z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextRunsRFC3339(t, tt.cronFormat, time.UTC, from, len(tt.want))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCronSchedule_TimeZones(t *testing.T) {
	stockholm := mustLoadLocation(t, "Europe/Stockholm")
	from := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		cronFormat string
		want       string
	}{
		{"schedule time zone", "0 3 * * *", "2026-10-09T03:00:00+02:00"},
		{"time zone prefix", "CRON_TZ=America/New_York 0 3 * * *", "2026-10-09T03:00:00-04:00"},
		{"time zone prefix with a Quartz day form", "TZ=America/New_York 0 3 * * 2#2", "2026-10-13T03:00:00-04:00"},
		{"Local means the schedule time zone, as in robfig/cron", "TZ=Local 0 3 * * *", "2026-10-09T03:00:00+02:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextRunsRFC3339(t, tt.cronFormat, stockholm, from, 1)
			assert.Equal(t, []string{tt.want}, got)
		})
	}
}

func TestParseCronSchedule_DSTChanges(t *testing.T) {
	tests := []struct {
		name       string
		zone       string
		cronFormat string
		from       string
		want       []string
	}{
		{
			name:       "a repeated hour runs once",
			zone:       "Europe/Stockholm",
			cronFormat: "30 2 * * 0L",
			from:       "2026-10-01T00:00:00+02:00",
			want:       []string{"2026-10-25T02:30:00+01:00", "2026-11-29T02:30:00+01:00"},
		},
		{
			name:       "clocks change at midnight",
			zone:       "Africa/Cairo",
			cronFormat: "30 23 * * 5L",
			from:       "2027-04-01T00:00:00+02:00",
			want:       []string{"2027-04-30T23:30:00+03:00", "2027-05-28T23:30:00+03:00"},
		},
		{
			name:       "clocks change at midnight on the matching day",
			zone:       "Asia/Beirut",
			cronFormat: "0 23 * * 0L",
			from:       "2027-03-01T00:00:00+02:00",
			want:       []string{"2027-03-28T23:00:00+03:00", "2027-04-25T23:00:00+03:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.zone+": "+tt.name, func(t *testing.T) {
			from, err := time.Parse(time.RFC3339, tt.from)
			require.NoError(t, err)

			got := nextRunsRFC3339(t, tt.cronFormat, mustLoadLocation(t, tt.zone), from, len(tt.want))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCronSchedule_NextIsAfterTheGivenTime(t *testing.T) {
	schedule, err := ParseCronSchedule("0 3 * * 2#2", time.UTC)
	require.NoError(t, err)

	run := time.Date(2026, 10, 13, 3, 0, 0, 0, time.UTC)

	assert.Equal(t, run, schedule.Next(run.Add(-time.Second)))
	assert.Equal(t, time.Date(2026, 11, 10, 3, 0, 0, 0, time.UTC), schedule.Next(run))
}

func TestParseCronSchedule_Errors(t *testing.T) {
	tests := []struct {
		name       string
		cronFormat string
		wantErr    string
	}{
		{"both day fields set", "0 3 1 * 2#2", "must be * or ?"},
		{"occurrence out of range", "0 3 * * 2#6", "must list weekdays"},
		{"zero-padded weekday", "0 3 * * 02#2", "must list weekdays"},
		{"range of weekdays", "0 3 * * 1-5#1", "must list weekdays"},
		{"list of month days", "0 3 1W,15W * *", "must be L, LW or a day"},
		{"days before the end of the month", "0 3 L-3 * *", "must be L, LW or a day"},
		{"reversed range", "5-2 3 * * 2#2", "beyond end of range"},
		{"wrap-around range", "0 3 * NOV-FEB 5L", "beyond end of range"},
		{"descriptor in a field", "@hourly * * * 2#2", "unrecognized descriptor"},
		{"six fields", "0 3 * * 2#2 2027", "expected exactly 5 fields"},
		{"time zone without schedule", "TZ=UTC", "missing schedule after TZ=UTC"},
		{"empty time zone without schedule", "CRON_TZ=", "missing schedule"},
		{"two time zones", "CRON_TZ=UTC TZ=UTC 0 3 * * *", "more than one time zone prefix"},
		{"unknown time zone", "CRON_TZ=Nowhere/City 0 3 * * *", "provided bad location Nowhere/City"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCronSchedule(tt.cronFormat, time.UTC)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestValidateCronFormat(t *testing.T) {
	tests := []struct {
		cronFormat string
		valid      bool
	}{
		{"* * * *", false},
		{"* * 1 * *", true},
		{"0 3 * * 2#2", true},
		{"@every 1h", true},
	}

	for _, tt := range tests {
		t.Run(tt.cronFormat, func(t *testing.T) {
			err := ValidateCronFormat(tt.cronFormat)
			assert.Equal(t, tt.valid, err == nil, "error: %v", err)
		})
	}
}

// The scheduler and the preview call Next on whatever was saved, so no input
// may panic or make Next go backwards.
func FuzzParseCronSchedule(f *testing.F) {
	seeds := []string{
		"0 3 * * 2#2", "30 23 * * 5L", "0 3 LW * *", "10,35 2 * * *", "@every 1s",
		"5-2 3 * * 2#2", "0 3 * 12-1 2#2", "@hourly * * * 2#2", "0 3 * * 02#2", "TZ=UTC",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	cairo := mustLoadLocation(f, "Africa/Cairo")
	from := time.Date(2027, 4, 29, 0, 0, 0, 0, cairo)

	f.Fuzz(func(t *testing.T, cronFormat string) {
		schedule, err := ParseCronSchedule(cronFormat, cairo)
		if err != nil {
			return
		}

		prev := from
		for _, run := range NextRuns(schedule, from, 3) {
			require.True(t, run.After(prev), "%s: %s is not after %s", cronFormat, run, prev)
			prev = run
		}
	})
}
