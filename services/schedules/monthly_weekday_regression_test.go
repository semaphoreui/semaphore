package schedules

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"
)

// TestDiffSweep_NoRegressionVsParseStandard is a large differential: for a broad
// generated set of expressions, the wrapper must make the SAME accept/reject
// decision as cron.ParseStandard and, when accepted, produce the SAME Next()
// over several steps. The only permitted difference is that where ParseStandard
// PANICS, the wrapper returns a clean error (never the reverse).
func TestDiffSweep_NoRegressionVsParseStandard(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping ~23k-spec differential sweep in -short mode")
	}
	safe := func(fn func() (cron.Schedule, error)) (s cron.Schedule, err error, panicked bool) {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		s, err = fn()
		return
	}

	minutes := []string{"*", "0", "59", "*/5", "0,30", "15-45", "1-59/2", "5"}
	hours := []string{"*", "0", "23", "*/2", "9-17", "0,12", "8"}
	doms := []string{"*", "1", "31", "*/10", "1-15", "15", "L-not"} // "L-not" is invalid -> both reject
	months := []string{"*", "1", "12", "JAN", "DEC", "*/3", "1-6", "JAN-MAR"}
	dows := []string{"*", "0", "6", "7", "MON", "SUN", "MON-FRI", "1-5", "0,6"}
	prefixes := []string{"", "TZ=UTC ", "CRON_TZ=America/New_York ", "CRON_TZ=Asia/Tokyo ", "CRON_TZ=Bad/Zone "}

	p := newCronParser()
	ref := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)

	var checked, accepted int
	for _, pfx := range prefixes {
		for _, mi := range minutes {
			for _, h := range hours {
				// Keep dom/month/dow varied but bounded to keep the sweep sane.
				for _, dm := range doms {
					for _, mo := range []string{months[0], months[3], months[6]} {
						for _, dw := range []string{dows[0], dows[3], dows[6], dows[8]} {
							spec := pfx + mi + " " + h + " " + dm + " " + mo + " " + dw
							checked++

							ws, werr, wpanic := safe(func() (cron.Schedule, error) { return p.Parse(spec) })
							ss, serr, spanic := safe(func() (cron.Schedule, error) { return cron.ParseStandard(spec) })

							require.False(t, wpanic, "wrapper PANICKED on %q", spec)

							wOK := werr == nil && !wpanic
							sOK := serr == nil && !spanic

							if spanic {
								// The only allowed asymmetry: std panics, wrapper errors cleanly.
								require.False(t, wOK, "wrapper accepted %q that ParseStandard panics on", spec)
								continue
							}

							require.Equalf(t, sOK, wOK,
								"accept/reject REGRESSION on %q: std=%v wrapper=%v (werr=%v serr=%v)",
								spec, sOK, wOK, werr, serr)

							if wOK && sOK {
								accepted++
								wt, st := ref, ref
								for i := 0; i < 6; i++ {
									wt = ws.Next(wt)
									st = ss.Next(st)
									require.Truef(t, wt.Equal(st),
										"Next DIVERGES on %q step %d: wrapper=%s std=%s", spec, i, wt, st)
								}
							}
						}
					}
				}
			}
		}
	}
	t.Logf("swept %d specs, %d accepted by both, zero Next divergences", checked, accepted)
}

// TestDiffSweep_Descriptors confirms robfig's own @-descriptors are unchanged by
// the wrapper (accept + identical Next), and only the new @monthly-weekday is
// additive.
func TestDiffSweep_Descriptors(t *testing.T) {
	p := newCronParser()
	ref := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	for _, d := range []string{"@yearly", "@annually", "@monthly", "@weekly", "@daily", "@midnight", "@hourly", "@every 1h30m", "@every 45m"} {
		ws, werr := p.Parse(d)
		ss, serr := cron.ParseStandard(d)
		require.NoError(t, werr, d)
		require.NoError(t, serr, d)
		wt, st := ref, ref
		for i := 0; i < 6; i++ {
			wt = ws.Next(wt)
			st = ss.Next(st)
			require.Truef(t, wt.Equal(st), "descriptor %q Next diverges at %d: %s vs %s", d, i, wt, st)
		}
	}
}
