package schedules

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupScheduleConfig points util.Config.Schedule at a fixed timezone for the
// duration of a test and restores the previous value afterwards.
func setupScheduleConfig(t *testing.T, tz string) {
	t.Helper()
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
	orig := util.Config.Schedule
	t.Cleanup(func() { util.Config.Schedule = orig })
	util.Config.Schedule = &util.ScheduleConfig{Timezone: tz}
}

// mustParse parses a descriptor and fails the test immediately if it is
// rejected — used only for specs that are expected to be valid.
func mustParse(t *testing.T, spec string) monthlyWeekdaySchedule {
	t.Helper()
	sched, err := parseMonthlyWeekday(spec, time.UTC)
	require.NoError(t, err)
	return sched
}

// TestMonthlyWeekday_ThirdWednesday walks a full year and asserts the third
// Wednesday of every month, at the configured time, in UTC.
func TestMonthlyWeekday_ThirdWednesday(t *testing.T) {
	sched := mustParse(t, "@monthly-weekday 3 wed at 09:00")

	// Third Wednesdays of 2026, verified against a calendar.
	expected := []string{
		"2026-01-21T09:00:00Z",
		"2026-02-18T09:00:00Z",
		"2026-03-18T09:00:00Z",
		"2026-04-15T09:00:00Z",
		"2026-05-20T09:00:00Z",
		"2026-06-17T09:00:00Z",
		"2026-07-15T09:00:00Z",
		"2026-08-19T09:00:00Z",
		"2026-09-16T09:00:00Z",
		"2026-10-21T09:00:00Z",
		"2026-11-18T09:00:00Z",
		"2026-12-16T09:00:00Z",
	}

	cur := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, want := range expected {
		next := sched.Next(cur)
		assert.Equal(t, want, next.Format(time.RFC3339))
		cur = next
	}
}

// TestMonthlyWeekday_FirstWednesdayAfterPatchTuesday is the headline case:
// second Tuesday + 1 day. The offset date is sometimes the second and sometimes
// the third Wednesday, which is exactly why no fixed-ordinal rule expresses it.
func TestMonthlyWeekday_FirstWednesdayAfterPatchTuesday(t *testing.T) {
	sched := mustParse(t, "@monthly-weekday 2 tue offset 1 at 03:00")

	expected := []string{
		"2026-01-14T03:00:00Z", // 2nd Tue Jan 13 -> Wed 14
		"2026-02-11T03:00:00Z", // 2nd Tue Feb 10 -> Wed 11
		"2026-03-11T03:00:00Z",
		"2026-04-15T03:00:00Z", // 2nd Tue Apr 14 -> Wed 15 (the 3rd Wednesday)
		"2026-05-13T03:00:00Z",
		"2026-06-10T03:00:00Z",
		"2026-07-15T03:00:00Z",
		"2026-08-12T03:00:00Z",
		"2026-09-09T03:00:00Z", // 2nd Tue Sep 8 -> Wed 9 (the 2nd Wednesday)
		"2026-10-14T03:00:00Z",
		"2026-11-11T03:00:00Z",
		"2026-12-09T03:00:00Z",
	}

	cur := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, want := range expected {
		next := sched.Next(cur)
		assert.Equal(t, want, next.Format(time.RFC3339))
		cur = next
	}
}

func TestMonthlyWeekday_LastFriday(t *testing.T) {
	sched := mustParse(t, "@monthly-weekday last fri at 22:30")

	tests := []struct {
		name string
		from string
		want string
	}{
		{"from month start", "2026-02-01T00:00:00Z", "2026-02-27T22:30:00Z"},
		{"31-day month", "2026-01-01T00:00:00Z", "2026-01-30T22:30:00Z"},
		{"last friday is the 31st", "2026-07-01T00:00:00Z", "2026-07-31T22:30:00Z"},
		{"leap february", "2028-02-01T00:00:00Z", "2028-02-25T22:30:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, err := time.Parse(time.RFC3339, tt.from)
			require.NoError(t, err)
			assert.Equal(t, tt.want, sched.Next(from).Format(time.RFC3339))
		})
	}
}

// TestMonthlyWeekday_FifthOccurrenceSkipsMonthsWithout confirms that a fifth
// weekday that does not exist in a month is skipped, not clamped.
func TestMonthlyWeekday_FifthOccurrenceSkipsMonthsWithout(t *testing.T) {
	sched := mustParse(t, "@monthly-weekday 5 sun at 12:00")

	// 2026: fifth Sundays occur in Mar(29), May(31), Aug(30), Nov(29).
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expected := []string{
		"2026-03-29T12:00:00Z",
		"2026-05-31T12:00:00Z",
		"2026-08-30T12:00:00Z",
		"2026-11-29T12:00:00Z",
	}
	for _, want := range expected {
		next := sched.Next(from)
		assert.Equal(t, want, next.Format(time.RFC3339))
		from = next
	}
}

// TestMonthlyWeekday_OffsetCrossesMonthBoundary checks that a positive offset
// from a late-month anchor lands in the following month and is still found when
// the reference time is already in that following month.
func TestMonthlyWeekday_OffsetCrossesMonthBoundary(t *testing.T) {
	// Last Sunday of Feb 2026 is the 22nd; +7 days -> Mar 1.
	sched := mustParse(t, "@monthly-weekday last sun offset 7 at 08:00")

	// Reference already past the anchor but before the offset date.
	from := time.Date(2026, time.February, 25, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2026-03-01T08:00:00Z", sched.Next(from).Format(time.RFC3339))
}

func TestMonthlyWeekday_NegativeOffset(t *testing.T) {
	// Second Tuesday of March 2026 is the 10th; offset -2 -> Sunday the 8th.
	sched := mustParse(t, "@monthly-weekday 2 tue offset -2 at 00:00")
	from := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2026-03-08T00:00:00Z", sched.Next(from).Format(time.RFC3339))
}

func TestMonthlyWeekday_DefaultsToMidnight(t *testing.T) {
	sched := mustParse(t, "@monthly-weekday 1 mon")
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2026-01-05T00:00:00Z", sched.Next(from).Format(time.RFC3339))
}

// A single-digit hour with a 2-digit minute is unambiguous and accepted.
func TestMonthlyWeekday_AcceptsSingleDigitHour(t *testing.T) {
	sched := mustParse(t, "@monthly-weekday 1 mon at 9:00")
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2026-01-05T09:00:00Z", sched.Next(from).Format(time.RFC3339))
}

func TestMonthlyWeekday_ClausesEitherOrder(t *testing.T) {
	a := mustParse(t, "@monthly-weekday 2 tue offset 1 at 03:00")
	b := mustParse(t, "@monthly-weekday 2 tue at 03:00 offset 1")
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, a.Next(from), b.Next(from))
}

// TestMonthlyWeekday_DST verifies the two properties that matter across a DST
// boundary: the wall-clock time of day is preserved, and the zone (hence the
// UTC instant) tracks DST. 09:00 is used because it is never inside a
// spring-forward gap, so the assertions are well-defined.
func TestMonthlyWeekday_DST(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// Second Sunday at 09:00 local. 2026 DST runs Mar 8 – Nov 1.
	sched, err := parseMonthlyWeekday("@monthly-weekday 2 sun at 09:00", ny)
	require.NoError(t, err)

	// January is EST; the second Sunday is the 11th.
	fromJan := time.Date(2026, time.January, 1, 0, 0, 0, 0, ny)
	nextJan := sched.Next(fromJan)
	assert.Equal(t, "2026-01-11T09:00:00", nextJan.Format("2006-01-02T15:04:05"))
	assert.Equal(t, "EST", nextJan.Format("MST"))

	// April is EDT; the second Sunday is the 12th — same wall clock, different
	// zone, so a different UTC instant.
	fromApr := time.Date(2026, time.April, 1, 0, 0, 0, 0, ny)
	nextApr := sched.Next(fromApr)
	assert.Equal(t, "2026-04-12T09:00:00", nextApr.Format("2006-01-02T15:04:05"))
	assert.Equal(t, "EDT", nextApr.Format("MST"))
}

// TestMonthlyWeekday_AmbientTimezone confirms a prefix-less schedule evaluates
// in the location of the time passed to Next, mirroring robfig/cron and letting
// the pool's WithLocation drive the effective zone.
func TestMonthlyWeekday_AmbientTimezone(t *testing.T) {
	sched := mustParseAmbient(t, "@monthly-weekday 3 wed at 09:00")

	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, ny)
	next := sched.Next(from)
	// 09:00 in New York, not UTC.
	assert.Equal(t, "2026-01-21T09:00:00", next.In(ny).Format("2006-01-02T15:04:05"))
	assert.Equal(t, "EST", next.In(ny).Format("MST"))
}

func mustParseAmbient(t *testing.T, spec string) monthlyWeekdaySchedule {
	t.Helper()
	sched, err := parseMonthlyWeekday(spec, nil)
	require.NoError(t, err)
	return sched
}

func TestParseMonthlyWeekday_Errors(t *testing.T) {
	tests := []struct {
		name string
		spec string
		msg  string
	}{
		{"missing weekday", "@monthly-weekday 3", "requires an ordinal and a weekday"},
		{"ordinal zero", "@monthly-weekday 0 wed", "ordinal must be 1..5 or 'last'"},
		{"ordinal too big", "@monthly-weekday 6 wed", "ordinal must be 1..5 or 'last'"},
		{"ordinal not a number", "@monthly-weekday x wed", "ordinal must be 1..5 or 'last'"},
		{"unknown weekday", "@monthly-weekday 3 funday", "unknown weekday"},
		{"offset without value", "@monthly-weekday 3 wed offset", "offset requires a number"},
		{"offset not a number", "@monthly-weekday 3 wed offset abc", "offset must be an integer"},
		{"offset out of range", "@monthly-weekday 3 wed offset 40", "within ±28 days"},
		{"at without value", "@monthly-weekday 3 wed at", "at requires a HH:MM"},
		{"time not HH:MM", "@monthly-weekday 3 wed at 9", "time must be HH:MM"},
		{"ambiguous single-digit minute", "@monthly-weekday 3 wed at 9:5", "time must be HH:MM"},
		{"single-digit minute padded hour", "@monthly-weekday 3 wed at 09:5", "time must be HH:MM"},
		{"time with seconds", "@monthly-weekday 3 wed at 09:00:00", "time must be HH:MM"},
		{"time with sign", "@monthly-weekday 3 wed at +9:00", "time must be HH:MM"},
		{"hour out of range", "@monthly-weekday 3 wed at 24:00", "hour must be 0..23"},
		{"minute out of range", "@monthly-weekday 3 wed at 09:60", "minute must be 0..59"},
		{"duplicate offset", "@monthly-weekday 3 wed offset 1 offset 2", "duplicate offset"},
		{"duplicate at", "@monthly-weekday 3 wed at 09:00 at 10:00", "duplicate at"},
		{"unexpected token", "@monthly-weekday 3 wed frequently", "unexpected token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseMonthlyWeekday(tt.spec, time.UTC)
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.msg)
		})
	}
}

func TestParseMonthlyWeekday_AcceptsWeekdaySpellings(t *testing.T) {
	spellings := map[string]time.Weekday{
		"sun": time.Sunday, "SUNDAY": time.Sunday,
		"Mon": time.Monday, "tue": time.Tuesday, "tues": time.Tuesday,
		"WED": time.Wednesday, "thurs": time.Thursday,
		"Fri": time.Friday, "saturday": time.Saturday,
	}
	for spelling, want := range spellings {
		t.Run(spelling, func(t *testing.T) {
			sched, err := parseMonthlyWeekday("@monthly-weekday 1 "+spelling, time.UTC)
			require.NoError(t, err)
			assert.Equal(t, want, sched.weekday)
		})
	}
}

// TestCronParser_DelegatesAndDetects verifies the wrapping parser routes
// descriptors to the monthly-weekday grammar, delegates ordinary cron and
// built-in descriptors to the standard parser, and honours a CRON_TZ prefix on
// a monthly-weekday descriptor.
func TestCronParser_DelegatesAndDetects(t *testing.T) {
	p := newCronParser()

	t.Run("monthly-weekday", func(t *testing.T) {
		sched, err := p.Parse("@monthly-weekday 3 wed at 09:00")
		require.NoError(t, err)
		_, ok := sched.(monthlyWeekdaySchedule)
		assert.True(t, ok, "expected a monthlyWeekdaySchedule")
	})

	t.Run("standard cron still works", func(t *testing.T) {
		sched, err := p.Parse("* * 1 * *")
		require.NoError(t, err)
		assert.NotNil(t, sched)
	})

	t.Run("builtin descriptor still works", func(t *testing.T) {
		sched, err := p.Parse("@hourly")
		require.NoError(t, err)
		assert.NotNil(t, sched)
	})

	t.Run("CRON_TZ prefix on descriptor pins the zone", func(t *testing.T) {
		sched, err := p.Parse("CRON_TZ=America/New_York @monthly-weekday 3 wed at 09:00")
		require.NoError(t, err)
		mw, ok := sched.(monthlyWeekdaySchedule)
		require.True(t, ok)
		require.NotNil(t, mw.loc)
		assert.Equal(t, "America/New_York", mw.loc.String())
	})

	t.Run("bad zone prefix is rejected", func(t *testing.T) {
		_, err := p.Parse("CRON_TZ=Bad/Zone @monthly-weekday 3 wed")
		require.Error(t, err)
	})

	t.Run("descriptor errors surface", func(t *testing.T) {
		_, err := p.Parse("@monthly-weekday 9 wed")
		require.Error(t, err)
		assert.ErrorContains(t, err, "ordinal")
	})
}

// TestValidateCronFormat_Extended confirms validation and firing share the one
// grammar: the descriptor is accepted, ordinary cron still is, and garbage is
// still rejected.
func TestValidateCronFormat_Extended(t *testing.T) {
	assert.NoError(t, ValidateCronFormat("@monthly-weekday 2 tue offset 1 at 03:00"))
	assert.NoError(t, ValidateCronFormat("* * 1 * *"))
	assert.Error(t, ValidateCronFormat("* * * *"))
	assert.Error(t, ValidateCronFormat("@monthly-weekday 3 funday"))
}

func TestNextRunTimes(t *testing.T) {
	setupScheduleConfig(t, "UTC")

	runs, err := NextRunTimes("@monthly-weekday 3 wed at 09:00", 3)
	require.NoError(t, err)
	require.Len(t, runs, 3)
	// Strictly increasing and all in the future.
	assert.True(t, runs[0].After(time.Now()))
	assert.True(t, runs[1].After(runs[0]))
	assert.True(t, runs[2].After(runs[1]))

	_, err = NextRunTimes("not a cron", 3)
	assert.Error(t, err)
}
