package schedules

import (
	"fmt"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Independent oracle for the monthly-weekday rule, used to property-test
// monthlyWeekdaySchedule.Next against a deliberately different algorithm.
//
// nthWeekdayOfMonth finds the day-of-month of the Nth (1..5) or last occurrence
// of weekday wd in (year, month), by *counting days independently* (no modulo
// arithmetic — deliberately different from the implementation). Returns ok=false
// when a positive ordinal does not occur that month.
// ---------------------------------------------------------------------------
func nthWeekdayOfMonth(year int, month time.Month, wd time.Weekday, ordinal int, loc *time.Location) (int, bool) {
	daysIn := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()

	var matches []int
	for d := 1; d <= daysIn; d++ {
		if time.Date(year, month, d, 0, 0, 0, 0, loc).Weekday() == wd {
			matches = append(matches, d)
		}
	}
	if len(matches) == 0 {
		return 0, false
	}
	if ordinal == ordinalLast {
		return matches[len(matches)-1], true
	}
	if ordinal < 1 || ordinal > len(matches) {
		return 0, false
	}
	return matches[ordinal-1], true
}

// oracleNext computes the true next activation strictly after t by enumerating
// every candidate anchor over a wide window (t-3 months .. t+16 months),
// applying the SAME fire formula the spec defines (anchor at H:M, then AddDate
// offset), and taking the minimum instant strictly after t.
func oracleNext(s monthlyWeekdaySchedule, t time.Time) time.Time {
	loc := s.loc
	if loc == nil {
		loc = t.Location()
	}
	t = t.In(loc)

	year, month := t.Year(), t.Month()
	// Two months before t through nine after: with |offset| <= 28 and monthly
	// occurrences, the true next is always inside this window.
	startYear, startMonth := year, month-2
	for startMonth < time.January {
		startMonth += 12
		startYear--
	}

	var best time.Time
	yy, mm := startYear, startMonth
	for i := 0; i < 12; i++ {
		if day, ok := nthWeekdayOfMonth(yy, mm, s.weekday, s.ordinal, loc); ok {
			anchor := time.Date(yy, mm, day, s.hour, s.minute, 0, 0, loc)
			fire := anchor.AddDate(0, 0, s.offsetDays)
			if fire.After(t) && (best.IsZero() || fire.Before(best)) {
				best = fire
			}
		}
		mm++
		if mm > time.December {
			mm = time.January
			yy++
		}
	}
	return best
}

// combos returns a representative grid of schedules.
func combos(loc *time.Location) []monthlyWeekdaySchedule {
	var out []monthlyWeekdaySchedule
	ordinals := []int{1, 2, 3, 4, 5, ordinalLast}
	weekdays := []time.Weekday{
		time.Sunday, time.Monday, time.Tuesday, time.Wednesday,
		time.Thursday, time.Friday, time.Saturday,
	}
	offsets := []int{-28, -15, -7, -1, 0, 1, 7, 15, 28}
	hm := [][2]int{{0, 0}, {9, 0}, {2, 30}, {23, 59}}
	for _, o := range ordinals {
		for _, w := range weekdays {
			for _, off := range offsets {
				h := hm[(o+int(w)+off)&3] // vary time deterministically
				out = append(out, monthlyWeekdaySchedule{
					ordinal: o, weekday: w, offsetDays: off,
					hour: h[0], minute: h[1], loc: loc,
				})
			}
		}
	}
	return out
}

// TestMonthlyWeekday_Prop_BruteForceOracle is a brute-force property check: for a large grid of schedules
// and many reference instants across several years, Next(t) must equal an
// independent oracle. Also feeds occurrence instants back in to exercise the
// strict-after boundary.
func TestMonthlyWeekday_Prop_BruteForceOracle(t *testing.T) {
	loc := time.UTC
	refs := []time.Time{}
	// Dense sampling across ~6 years, including odd hours/minutes and
	// month/year boundaries.
	for _, y := range []int{2026, 2027, 2028} { // includes a leap year
		for _, m := range []time.Month{time.February, time.December} {
			for _, d := range []int{1, 15, 28} {
				for _, hh := range []int{0, 23} {
					refs = append(refs, time.Date(y, m, d, hh, 30, 0, 0, loc))
				}
			}
		}
	}

	for _, s := range combos(loc) {
		for _, ref := range refs {
			got := s.Next(ref)
			want := oracleNext(s, ref)
			require.Truef(t, got.Equal(want), "MISMATCH desc=%s ref=%s got=%s want=%s",
				descOf(s), ref.Format(time.RFC3339), fmtT(got), fmtT(want))
			// Strict-after: feed the result back; must advance and match oracle.
			if !got.IsZero() {
				assert.True(t, got.After(ref), "Next not strictly after ref: %s ref=%s got=%s", descOf(s), ref, got)
				got2 := s.Next(got)
				want2 := oracleNext(s, got)
				require.Truef(t, got2.Equal(want2), "MISMATCH(2nd) desc=%s from=%s got=%s want=%s",
					descOf(s), fmtT(got), fmtT(got2), fmtT(want2))
			}
		}
	}
}

// TestMonthlyWeekday_Prop_OneMonthBackSufficiency directly stresses the "start one month
// before t" assumption with the worst-case offsets and last/5th ordinals at
// year rollover.
func TestMonthlyWeekday_Prop_OneMonthBackSufficiency(t *testing.T) {
	loc := time.UTC
	worst := []monthlyWeekdaySchedule{
		{ordinal: ordinalLast, weekday: time.Friday, offsetDays: 28, hour: 23, minute: 59, loc: loc},
		{ordinal: ordinalLast, weekday: time.Saturday, offsetDays: -28, hour: 0, minute: 0, loc: loc},
		{ordinal: 5, weekday: time.Sunday, offsetDays: 28, hour: 23, minute: 59, loc: loc},
		{ordinal: 5, weekday: time.Monday, offsetDays: -28, hour: 0, minute: 0, loc: loc},
		{ordinal: 1, weekday: time.Sunday, offsetDays: -28, hour: 0, minute: 0, loc: loc},
	}
	// Every hour across two full years — will catch any boundary miss.
	for _, s := range worst {
		cur := time.Date(2027, time.January, 1, 0, 0, 0, 0, loc)
		end := time.Date(2029, time.January, 1, 0, 0, 0, 0, loc)
		for cur.Before(end) {
			got := s.Next(cur)
			want := oracleNext(s, cur)
			require.Truef(t, got.Equal(want), "MISMATCH desc=%s ref=%s got=%s want=%s",
				descOf(s), cur.Format(time.RFC3339), fmtT(got), fmtT(want))
			cur = cur.Add(6 * time.Hour)
		}
	}
}

// TestMonthlyWeekday_Prop_DSTZones checks DST zones: monotonicity + oracle agreement in
// tricky zones (NY: 1h DST; Chatham: :45 offset + DST; Lord Howe: 30-min DST),
// including anchors deliberately at wall times inside spring-forward gaps.
func TestMonthlyWeekday_Prop_DSTZones(t *testing.T) {
	for _, zone := range []string{"America/New_York", "Pacific/Chatham", "Australia/Lord_Howe"} {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			t.Logf("zone %s unavailable, skipping: %v", zone, err)
			continue
		}
		t.Run(zone, func(t *testing.T) {
			// Include gap-prone wall times: 02:30 (NY/LordHowe), 02:45/03:45 (Chatham).
			schedules := []monthlyWeekdaySchedule{
				{ordinal: 2, weekday: time.Sunday, hour: 2, minute: 30, loc: loc},
				{ordinal: 1, weekday: time.Sunday, hour: 2, minute: 45, loc: loc},
				{ordinal: ordinalLast, weekday: time.Sunday, hour: 2, minute: 30, offsetDays: 1, loc: loc},
				{ordinal: 3, weekday: time.Wednesday, hour: 9, minute: 0, offsetDays: -1, loc: loc},
			}
			for _, s := range schedules {
				prev := time.Date(2024, time.January, 1, 0, 0, 0, 0, loc)
				for i := 0; i < 120; i++ { // 10 years of monthly fires
					next := s.Next(prev)
					require.False(t, next.IsZero(), "unexpected zero: %s from %s", descOf(s), prev)
					assert.True(t, next.After(prev),
						"NON-MONOTONIC desc=%s prev=%s next=%s", descOf(s), fmtT(prev), fmtT(next))
					want := oracleNext(s, prev)
					require.Truef(t, next.Equal(want), "MISMATCH desc=%s prev=%s got=%s want=%s",
						descOf(s), fmtT(prev), fmtT(next), fmtT(want))
					prev = next
				}
			}
		})
	}
}

// TestMonthlyWeekday_Prop_NeverSilentlyNeverFires confirms no valid descriptor returns the
// zero time (a schedule that silently never fires) within a long horizon — i.e.
// the 60-month cap is never hit for a valid weekday rule.
func TestMonthlyWeekday_Prop_NeverSilentlyNeverFires(t *testing.T) {
	loc := time.UTC
	for _, s := range combos(loc) {
		// Sample a few starting points per year across a decade.
		for y := 2024; y <= 2034; y++ {
			from := time.Date(y, time.February, 27, 12, 0, 0, 0, loc)
			got := s.Next(from)
			assert.False(t, got.IsZero(), "valid descriptor returned zero (never fires): %s from %s", descOf(s), from)
		}
	}
}

// ---- end-to-end through the real cron pool ----------------------------------

func TestMonthlyWeekday_Prop_PoolEndToEnd(t *testing.T) {
	loc := time.UTC
	c := cron.New(cron.WithLocation(loc), cron.WithParser(newCronParser()))

	id, err := c.AddJob("@monthly-weekday 3 wed at 09:00", cron.FuncJob(func() {}))
	require.NoError(t, err, "pool must accept the descriptor")
	entry := c.Entry(id)
	require.NotNil(t, entry.Schedule)
	next := entry.Schedule.Next(time.Date(2026, time.January, 1, 0, 0, 0, 0, loc))
	assert.Equal(t, "2026-01-21T09:00:00Z", next.Format(time.RFC3339))

	_, err = c.AddJob("@monthly-weekday 9 wed", cron.FuncJob(func() {}))
	assert.Error(t, err, "pool must reject an invalid descriptor")

	// CRON_TZ-pinned via the pool.
	id2, err := c.AddJob("CRON_TZ=America/New_York @monthly-weekday 2 sun at 09:00", cron.FuncJob(func() {}))
	require.NoError(t, err)
	ny, _ := time.LoadLocation("America/New_York")
	n2 := c.Entry(id2).Schedule.Next(time.Date(2026, time.January, 1, 0, 0, 0, 0, loc))
	assert.Equal(t, "EST", n2.In(ny).Format("MST"))
	assert.Equal(t, "2026-01-11T09:00:00", n2.In(ny).Format("2006-01-02T15:04:05"))
}

// TestMonthlyWeekday_Prop_ValidatePreviewMatchesFire checks that the validate preview matches firing: the validate preview
// (NextRunTimes, based on util.Config.Schedule.Timezone) must produce the same
// instants the pool would fire, for both a prefix-less descriptor (ambient =
// config TZ) and a CRON_TZ-pinned one (overrides config TZ).
func TestMonthlyWeekday_Prop_ValidatePreviewMatchesFire(t *testing.T) {
	setupScheduleConfig(t, "America/New_York") // non-UTC config to expose ambient bugs
	cfgLoc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	specs := []string{
		"@monthly-weekday 3 wed at 09:00",                    // prefix-less -> ambient/config TZ
		"CRON_TZ=Asia/Tokyo @monthly-weekday 2 mon at 06:30", // pinned, overrides config
		"CRON_TZ=America/New_York @monthly-weekday last fri at 22:30",
	}
	for _, spec := range specs {
		t.Run(spec, func(t *testing.T) {
			// Preview path (what the validate endpoint returns).
			preview, err := NextRunTimes(spec, 5)
			require.NoError(t, err)
			require.NotEmpty(t, preview)

			// Fire path: exactly how the pool evaluates it. The pool passes
			// time.Now().In(config TZ) into Schedule.Next.
			sched, err := newCronParser().Parse(spec)
			require.NoError(t, err)
			base := time.Now().In(cfgLoc)
			cur := base
			for i, want := range preview {
				cur = sched.Next(cur)
				assert.Truef(t, cur.Equal(want),
					"preview[%d]=%s != fire=%s for %q", i, fmtT(want), fmtT(cur), spec)
			}
		})
	}
}

// ---- delegation is a pure superset of ParseStandard ------------------------

func safeParse(fn func() (cron.Schedule, error)) (sched cron.Schedule, err error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	sched, err = fn()
	return
}

func TestMonthlyWeekday_Prop_DelegationParity(t *testing.T) {
	p := newCronParser()
	specs := []string{
		"* * * * *",
		"0 0 1 * *",
		"*/5 * * * *",
		"@hourly", "@daily", "@midnight", "@weekly", "@monthly", "@yearly", "@annually",
		"@every 1h30m", "@every 90m", "@every 0s",
		"CRON_TZ=UTC * * * * *",
		"CRON_TZ=America/New_York 0 9 * * 1",
		"TZ=Europe/Stockholm 0 0 * * *",
		"CRON_TZ=Bad/Zone * * * * *",
		"* * * *",     // too few
		"* * * * * *", // too many (standard = 5)
		"@bogus",
		"",
		"   ",
		"@every notaduration",
		"60 * * * *", // minute out of range
		"* 24 * * *", // hour out of range
		"* * 0 * *",  // dom below min
		"1-5/0 * * * *",
		"0 0 * * MON-FRI",
		"0 0 * * 7", // dow 7
		"\t* * * * *\t",
		"  * * * * *  ",
		"CRON_TZ=UTC\t* * * * *", // tab after tz prefix
	}

	type div struct {
		spec                string
		featOK, stdOK       bool
		featErr, stdErr     string
		featPanic, stdPanic bool
	}
	var divergences []div

	for _, spec := range specs {
		fs, fe, fp := safeParse(func() (cron.Schedule, error) { return p.Parse(spec) })
		ss, se, sp := safeParse(func() (cron.Schedule, error) { return cron.ParseStandard(spec) })

		featOK := fe == nil && !fp
		stdOK := se == nil && !sp

		// Record any difference in accept/reject or any panic.
		if featOK != stdOK || fp || sp {
			d := div{spec: spec, featOK: featOK, stdOK: stdOK, featPanic: fp, stdPanic: sp}
			if fe != nil {
				d.featErr = fe.Error()
			}
			if se != nil {
				d.stdErr = se.Error()
			}
			divergences = append(divergences, d)
		}

		// For specs both accept, compare Next over several steps.
		if featOK && stdOK {
			cur := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
			for i := 0; i < 5; i++ {
				fn := fs.Next(cur)
				sn := ss.Next(cur)
				assert.Truef(t, fn.Equal(sn),
					"Next diverges for %q at step %d: feat=%s std=%s", spec, i, fn, sn)
				cur = sn
				if cur.IsZero() {
					break
				}
			}
		}
	}

	for _, d := range divergences {
		t.Logf("DIVERGENCE spec=%q feat(ok=%v panic=%v err=%q) std(ok=%v panic=%v err=%q)",
			d.spec, d.featOK, d.featPanic, d.featErr, d.stdOK, d.stdPanic, d.stdErr)
	}

	// The only *acceptable* divergences are ones where the feature is SAFER:
	// it returns a clean error where robfig ParseStandard panics. Any case where
	// feature accepts and std rejects (or vice versa) WITHOUT a std panic is a
	// real regression.
	for _, d := range divergences {
		if d.stdPanic {
			assert.False(t, d.featPanic, "feature also panics on %q (should be clean error)", d.spec)
			assert.False(t, d.featOK, "feature accepts %q that std panics on", d.spec)
			continue
		}
		assert.Failf(t, "accept/reject regression", "on %q: feat=%v std=%v", d.spec, d.featOK, d.stdOK)
	}
}

// TestMonthlyWeekday_Prop_OldValidatePanicked demonstrates the pre-change ValidateCronFormat
// (cron.ParseStandard) PANICS on a TZ prefix with no following field, while the
// new parser returns a clean error — a latent DoS in the old validate path.
func TestMonthlyWeekday_Prop_OldValidatePanicked(t *testing.T) {
	_, _, oldPanic := safeParse(func() (cron.Schedule, error) { return cron.ParseStandard("TZ=UTC") })
	assert.True(t, oldPanic, "expected robfig ParseStandard to panic on \"TZ=UTC\"")

	// New parser (backs both validate and fire) must NOT panic and must reject.
	_, newErr, newPanic := safeParse(func() (cron.Schedule, error) { return newCronParser().Parse("TZ=UTC") })
	assert.False(t, newPanic, "new parser must not panic on \"TZ=UTC\"")
	assert.Error(t, newErr, "new parser must reject \"TZ=UTC\"")

	// And ValidateCronFormat itself must be panic-free.
	assert.NotPanics(t, func() { _ = ValidateCronFormat("TZ=UTC") })
	assert.NotPanics(t, func() { _ = ValidateCronFormat("CRON_TZ=Europe/Paris") })
}

// ---- grammar fuzzing: no panic, only in-range accepts ----------------------

func TestMonthlyWeekday_Prop_GrammarNoPanicAndClassify(t *testing.T) {
	cases := []string{
		"@monthly-weekday 99999999999999999999 wed",
		"@monthly-weekday 3 wed offset 99999999999999999999",
		"@monthly-weekday -3 wed",
		"@monthly-weekday +3 wed",
		"@monthly-weekday 3 wed at 9:5",
		"@monthly-weekday 3 wed at 09:5",
		"@monthly-weekday 3 wed at 9:05",
		"@monthly-weekday 3 wed at 24:00",
		"@monthly-weekday 3 wed at 23:60",
		"@monthly-weekday 3 wed at 23:59",
		"@monthly-weekday 3 wed offset -0",
		"@monthly-weekday 3 wed offset +5",
		"@monthly-weekday 3 WED AT 09:00",
		"@MONTHLY-WEEKDAY 3 wed",
		"@monthly-weekday 3 wed offset 1 at 09:00 extra",
		"@monthly-weekday 3 wed at 09:00 offset 1 offset 2",
		"@monthly-weekday 3 wed at :00",
		"@monthly-weekday 3 wed at 09:",
		"@monthly-weekday 3 wed at 09:00:00",
		"@monthly-weekday    3    wed",       // extra spaces
		"@monthly-weekday\t3\twed",           // tabs
		"@monthly-weekday 3 wed offset  2",   // double space before value
		"@monthly-weekday 3 wed offset",      // offset no value
		"@monthly-weekday 3 wed at",          // at no value
		"@monthly-weekday 3 wed offset 1 at", // trailing at no value
		"@monthly-weekday ３ wed",             // fullwidth digit
		"@monthly-weekday 3 wed at ０９:００",    // fullwidth time
		"@monthly-weekday 5 sun offset 28 at 23:59",
		"@monthly-weekday last sat offset -28",
		"@monthly-weekday 0x3 wed",
		"@monthly-weekday 3 wed offset 0x2",
	}
	for _, spec := range cases {
		spec := spec
		t.Run(spec, func(t *testing.T) {
			var (
				sched monthlyWeekdaySchedule
				err   error
			)
			assert.NotPanics(t, func() { sched, err = parseMonthlyWeekday(spec, time.UTC) },
				"parseMonthlyWeekday panicked on %q", spec)
			// Also route through the wrapping parser to ensure no panic there.
			assert.NotPanics(t, func() { _, _ = newCronParser().Parse(spec) })
			if err == nil {
				// If accepted, it must produce sane, evaluable schedule fields.
				assert.True(t, sched.ordinal == ordinalLast || (sched.ordinal >= 1 && sched.ordinal <= 5), "bad ordinal %d for %q", sched.ordinal, spec)
				assert.True(t, sched.hour >= 0 && sched.hour <= 23, "bad hour")
				assert.True(t, sched.minute >= 0 && sched.minute <= 59, "bad minute")
				assert.True(t, sched.offsetDays >= -28 && sched.offsetDays <= 28, "bad offset")
				// Evaluating must not panic and must yield a non-zero next.
				var n time.Time
				assert.NotPanics(t, func() { n = sched.Next(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) })
				assert.False(t, n.IsZero(), "accepted spec %q never fires", spec)
				t.Logf("ACCEPTED %q -> ord=%d wd=%v off=%d %02d:%02d", spec, sched.ordinal, sched.weekday, sched.offsetDays, sched.hour, sched.minute)
			} else {
				t.Logf("rejected %q: %v", spec, err)
			}
		})
	}
}

// ---- helpers --------------------------------------------------------------

func descOf(s monthlyWeekdaySchedule) string {
	ord := fmt.Sprintf("%d", s.ordinal)
	if s.ordinal == ordinalLast {
		ord = "last"
	}
	z := "ambient"
	if s.loc != nil {
		z = s.loc.String()
	}
	return fmt.Sprintf("@mw ord=%s wd=%v off=%d %02d:%02d [%s]", ord, s.weekday, s.offsetDays, s.hour, s.minute, z)
}

func fmtT(t time.Time) string {
	if t.IsZero() {
		return "<zero>"
	}
	return t.Format("2006-01-02T15:04:05Z07:00")
}
