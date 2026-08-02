package schedules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/semaphoreui/semaphore/util"
)

// monthlyWeekdayKeyword is the descriptor that selects a "Nth weekday of the
// month, plus an optional day offset" schedule. It is intentionally namespaced
// so it can never collide with a standard cron expression or one of
// robfig/cron's built-in descriptors (@monthly, @weekly, ...).
const monthlyWeekdayKeyword = "@monthly-weekday"

// ordinalLast marks the "last occurrence of the weekday in the month" variant.
// Regular ordinals are the positive integers 1..5.
const ordinalLast = -1

// maxWeekdayOffsetDays bounds the day offset. An offset larger than a month is
// meaningless (it would be better expressed as a different ordinal/weekday) and
// keeping it small guarantees the Next scan only has to look one month back to
// catch an occurrence that spilled across a month boundary.
const maxWeekdayOffsetDays = 28

// nextScanMonths bounds how far Next scans before giving up, matching
// robfig/cron's own five-year cap. It is only ever reached for the pathological
// case of an ordinal that never occurs, which cannot happen for a valid weekday.
const nextScanMonths = 5 * 12

// monthlyWeekdaySchedule fires on the Nth (or last) occurrence of a weekday in
// each month, shifted by offsetDays calendar days, at hour:minute. It
// implements cron.Schedule and is immutable, so it is safe to share and to
// evaluate concurrently on every node of an HA cluster.
type monthlyWeekdaySchedule struct {
	ordinal    int          // 1..5, or ordinalLast
	weekday    time.Weekday // Sunday..Saturday
	offsetDays int          // calendar-day shift applied to the anchor, [-28, 28]
	hour       int          // 0..23
	minute     int          // 0..59

	// loc pins the evaluation timezone when the descriptor carried a
	// CRON_TZ=/TZ= prefix. When nil the schedule is evaluated in the ambient
	// location of the time passed to Next — exactly how robfig/cron treats a
	// prefix-less expression (time.Local sentinel), which lets the pool's
	// WithLocation continue to drive the effective timezone.
	loc *time.Location
}

// Next returns the first activation strictly after t, or the zero time if none
// is found within the scan bound. It satisfies cron.Schedule.
func (s monthlyWeekdaySchedule) Next(t time.Time) time.Time {
	loc := s.loc
	if loc == nil {
		loc = t.Location()
	}
	t = t.In(loc)

	// Start one month before t. A positive offset applied to the previous
	// month's occurrence can land after t, and occurrences are strictly
	// increasing month over month, so scanning from the previous month and
	// returning the first occurrence after t yields the true next activation.
	year, month := t.Year(), t.Month()
	month--
	if month < time.January {
		month = time.December
		year--
	}

	for i := 0; i < nextScanMonths; i++ {
		if occ, ok := s.occurrenceIn(year, month, loc); ok && occ.After(t) {
			return occ
		}
		month++
		if month > time.December {
			month = time.January
			year++
		}
	}

	return time.Time{}
}

// occurrenceIn computes the activation instant for a specific month. ok is
// false when the requested ordinal does not occur that month (for example, a
// month with no fifth Friday).
func (s monthlyWeekdaySchedule) occurrenceIn(year int, month time.Month, loc *time.Location) (time.Time, bool) {
	last := daysInMonth(year, month, loc)

	var day int
	if s.ordinal == ordinalLast {
		lastDate := time.Date(year, month, last, 0, 0, 0, 0, loc)
		back := (int(lastDate.Weekday()) - int(s.weekday) + 7) % 7
		day = last - back
	} else {
		first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
		forward := (int(s.weekday) - int(first.Weekday()) + 7) % 7
		day = 1 + forward + (s.ordinal-1)*7
		if day > last {
			return time.Time{}, false
		}
	}

	anchor := time.Date(year, month, day, s.hour, s.minute, 0, 0, loc)
	// AddDate normalises across month and DST boundaries while preserving the
	// wall-clock time of day, so "second Tuesday at 03:00, offset 1 day" fires
	// on the following Wednesday at 03:00.
	return anchor.AddDate(0, 0, s.offsetDays), true
}

// daysInMonth returns the number of days in the given month. Day 0 of the
// following month is the last day of this one; time.Date normalises a December
// query (month+1 == 13) into the next year.
func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

// cronParser parses Semaphore's monthly-weekday descriptor and delegates every
// other expression to a standard robfig/cron parser. It implements
// cron.ScheduleParser so it can back both the scheduling pool (via
// cron.WithParser) and ValidateCronFormat, guaranteeing that what validates is
// exactly what fires.
type cronParser struct {
	standard cron.Parser
}

// newCronParser builds a parser whose fallback accepts exactly the same
// expressions as cron.ParseStandard (five fields plus @-descriptors). It holds
// no package-level state, so it can be constructed wherever it is needed.
func newCronParser() cronParser {
	return cronParser{
		standard: cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		),
	}
}

// Parse returns a cron.Schedule for spec. A monthly-weekday descriptor (with an
// optional CRON_TZ=/TZ= prefix) is parsed here; anything else is handed to the
// standard parser unchanged, including its own prefix handling.
func (p cronParser) Parse(spec string) (cron.Schedule, error) {
	body := strings.TrimSpace(spec)

	var loc *time.Location
	if strings.HasPrefix(body, "TZ=") || strings.HasPrefix(body, "CRON_TZ=") {
		eq := strings.Index(body, "=")
		sp := strings.IndexAny(body, " \t")
		if sp < 0 {
			return nil, fmt.Errorf("provided bad location: %s", body)
		}
		z, err := time.LoadLocation(strings.TrimSpace(body[eq+1 : sp]))
		if err != nil {
			return nil, err
		}
		loc = z
		body = strings.TrimSpace(body[sp+1:])
	}

	if isMonthlyWeekdaySpec(body) {
		return parseMonthlyWeekday(body, loc)
	}

	// Delegate the original spec so the standard parser applies its own,
	// identical prefix handling — no double interpretation.
	return p.standard.Parse(spec)
}

// isMonthlyWeekdaySpec reports whether body (already stripped of any timezone
// prefix) is a monthly-weekday descriptor.
func isMonthlyWeekdaySpec(body string) bool {
	fields := strings.Fields(body)
	return len(fields) > 0 && strings.EqualFold(fields[0], monthlyWeekdayKeyword)
}

// parseMonthlyWeekday turns a descriptor body into a monthlyWeekdaySchedule.
// The grammar is:
//
//	@monthly-weekday <ordinal> <weekday> [offset <days>] [at <HH:MM>]
//
// The optional "offset" and "at" clauses may appear in either order but at most
// once each. Every malformed input yields a descriptive error and never a
// partially-built schedule.
func parseMonthlyWeekday(body string, loc *time.Location) (monthlyWeekdaySchedule, error) {
	var zero monthlyWeekdaySchedule

	fields := strings.Fields(body)
	if len(fields) < 3 {
		return zero, fmt.Errorf(
			"%s requires an ordinal and a weekday, e.g. %q",
			monthlyWeekdayKeyword, monthlyWeekdayKeyword+" 3 wed at 09:00")
	}

	ordinal, err := parseOrdinal(fields[1])
	if err != nil {
		return zero, err
	}

	weekday, err := parseWeekday(fields[2])
	if err != nil {
		return zero, err
	}

	s := monthlyWeekdaySchedule{
		ordinal: ordinal,
		weekday: weekday,
		loc:     loc,
	}

	seenOffset := false
	seenAt := false
	for i := 3; i < len(fields); {
		switch strings.ToLower(fields[i]) {
		case "offset":
			if seenOffset {
				return zero, fmt.Errorf("%s: duplicate offset clause", monthlyWeekdayKeyword)
			}
			if i+1 >= len(fields) {
				return zero, fmt.Errorf("%s: offset requires a number of days", monthlyWeekdayKeyword)
			}
			s.offsetDays, err = parseOffset(fields[i+1])
			if err != nil {
				return zero, err
			}
			seenOffset = true
			i += 2
		case "at":
			if seenAt {
				return zero, fmt.Errorf("%s: duplicate at clause", monthlyWeekdayKeyword)
			}
			if i+1 >= len(fields) {
				return zero, fmt.Errorf("%s: at requires a HH:MM time", monthlyWeekdayKeyword)
			}
			s.hour, s.minute, err = parseHourMinute(fields[i+1])
			if err != nil {
				return zero, err
			}
			seenAt = true
			i += 2
		default:
			return zero, fmt.Errorf(
				"%s: unexpected token %q (expected 'offset <days>' or 'at <HH:MM>')",
				monthlyWeekdayKeyword, fields[i])
		}
	}

	return s, nil
}

// parseOrdinal accepts 1..5 or the word "last".
func parseOrdinal(s string) (int, error) {
	if strings.EqualFold(s, "last") {
		return ordinalLast, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 5 {
		return 0, fmt.Errorf("%s: ordinal must be 1..5 or 'last', got %q", monthlyWeekdayKeyword, s)
	}
	return n, nil
}

// parseWeekday accepts every common spelling of a weekday. Numeric weekdays are
// deliberately not accepted: the 0=Sunday vs 1=Monday ambiguity between cron
// dialects is exactly the kind of silent misconfiguration this feature exists
// to avoid.
func parseWeekday(s string) (time.Weekday, error) {
	switch strings.ToLower(s) {
	case "sun", "sunday":
		return time.Sunday, nil
	case "mon", "monday":
		return time.Monday, nil
	case "tue", "tues", "tuesday":
		return time.Tuesday, nil
	case "wed", "weds", "wednesday":
		return time.Wednesday, nil
	case "thu", "thur", "thurs", "thursday":
		return time.Thursday, nil
	case "fri", "friday":
		return time.Friday, nil
	case "sat", "saturday":
		return time.Saturday, nil
	default:
		return 0, fmt.Errorf("%s: unknown weekday %q (use sun..sat)", monthlyWeekdayKeyword, s)
	}
}

func parseOffset(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%s: offset must be an integer number of days, got %q", monthlyWeekdayKeyword, s)
	}
	if n < -maxWeekdayOffsetDays || n > maxWeekdayOffsetDays {
		return 0, fmt.Errorf("%s: offset must be within ±%d days, got %d", monthlyWeekdayKeyword, maxWeekdayOffsetDays, n)
	}
	return n, nil
}

func parseHourMinute(s string) (hour int, minute int, err error) {
	h, m, ok := strings.Cut(s, ":")
	// Require a canonical 24-hour HH:MM: a 1-2 digit hour and an exactly
	// 2-digit minute, digits only. "9:5" is ambiguous (09:05 vs 09:50) and a
	// signed or non-ASCII value is not a clock time, so all are rejected rather
	// than silently reinterpreted.
	if !ok || !isDecimal(h) || len(h) < 1 || len(h) > 2 || !isDecimal(m) || len(m) != 2 {
		return 0, 0, fmt.Errorf("%s: time must be HH:MM (24-hour), got %q", monthlyWeekdayKeyword, s)
	}
	hour, _ = strconv.Atoi(h)
	if hour > 23 {
		return 0, 0, fmt.Errorf("%s: hour must be 0..23, got %q", monthlyWeekdayKeyword, s)
	}
	minute, _ = strconv.Atoi(m)
	if minute > 59 {
		return 0, 0, fmt.Errorf("%s: minute must be 0..59, got %q", monthlyWeekdayKeyword, s)
	}
	return hour, minute, nil
}

// isDecimal reports whether s is a non-empty run of ASCII decimal digits (no
// sign, no whitespace, no non-ASCII digits).
func isDecimal(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// NextRunTimes returns up to count upcoming activations for a cron format,
// starting from now. It powers the schedule form's "Next run time" preview,
// which is authoritative for descriptors the frontend cron parser cannot
// evaluate. The base timezone matches the scheduling pool: an explicit
// CRON_TZ=/TZ= prefix wins, otherwise util.Config.Schedule.Timezone (UTC by
// default) is used.
func NextRunTimes(cronFormat string, count int) ([]time.Time, error) {
	schedule, err := newCronParser().Parse(cronFormat)
	if err != nil {
		return nil, err
	}

	t := time.Now().In(scheduleLocation())
	runs := make([]time.Time, 0, count)
	for i := 0; i < count; i++ {
		t = schedule.Next(t)
		if t.IsZero() {
			break
		}
		runs = append(runs, t)
	}
	return runs, nil
}

// scheduleLocation resolves the pool's base timezone, falling back to UTC when
// the config is absent or names an unknown zone.
func scheduleLocation() *time.Location {
	if util.Config != nil && util.Config.Schedule != nil {
		if loc, err := time.LoadLocation(util.Config.Schedule.Timezone); err == nil {
			return loc
		}
	}
	return time.UTC
}
