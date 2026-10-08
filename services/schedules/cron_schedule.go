package schedules

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/cronexpr"
	"github.com/robfig/cron/v3"
	"github.com/semaphoreui/semaphore/util"
)

// MaxOffsetDays is how far a cron schedule's runs may be moved, either way.
const MaxOffsetDays = 31

// maxClockChange is more than any DST change moves clocks: the largest in use
// is two hours (Antarctica/Troll).
const maxClockChange = 3 * time.Hour

// cronSchedule evaluates a cron_format in one time zone and moves every run by
// offsetDays calendar days.
type cronSchedule struct {
	expr         cron.Schedule
	location     *time.Location
	offsetDays   int
	useWallClock bool
}

func (s cronSchedule) Next(t time.Time) time.Time {
	t = t.In(s.location)

	if !s.useWallClock {
		return s.expr.Next(t)
	}

	// The expression runs on the wall clock in UTC, where no hour is skipped or
	// repeated, and each run is then placed in the schedule's zone. A DST change
	// can reorder runs close to it, so the search starts early and continues a
	// little past the first run found.
	from := wallClock(t).Add(-maxClockChange).AddDate(0, 0, -s.offsetDays)

	var next time.Time

	for run, prev := s.expr.Next(from), from; run.After(prev); run, prev = s.expr.Next(run), run {
		moved := run.AddDate(0, 0, s.offsetDays)

		if !next.IsZero() && moved.After(wallClock(next).Add(maxClockChange)) {
			break
		}

		if at := inZone(moved, s.location); at.After(t) && (next.IsZero() || at.Before(next)) {
			next = at
		}
	}

	return next
}

// wallClock returns the date and time of day of t as a time in UTC.
func wallClock(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
}

// inZone returns the time in loc with the date and time of day of wall. As with
// time.Date, a time skipped by DST moves past the change, and a time that occurs
// twice is placed at one of its occurrences.
func inZone(wall time.Time, loc *time.Location) time.Time {
	return time.Date(wall.Year(), wall.Month(), wall.Day(), wall.Hour(), wall.Minute(), wall.Second(), wall.Nanosecond(), loc)
}

// Location returns the time zone of cron formats without a TZ prefix.
func Location() (*time.Location, error) {
	return time.LoadLocation(util.Config.Schedule.Timezone)
}

// ParseCronSchedule parses a cron_format whose runs are moved by offsetDays.
// A format without a TZ prefix is evaluated in loc.
func ParseCronSchedule(cronFormat string, offsetDays int, loc *time.Location) (cron.Schedule, error) {
	if offsetDays < -MaxOffsetDays || offsetDays > MaxOffsetDays {
		return nil, fmt.Errorf("offset must be between %d and %d days", -MaxOffsetDays, MaxOffsetDays)
	}

	tz, expr, err := splitTimezone(cronFormat)
	if err != nil {
		return nil, err
	}

	if tz != nil {
		loc = tz
	}

	parsed, err := parseExpression(expr)
	if err != nil {
		return nil, err
	}

	if _, every := parsed.(cron.ConstantDelaySchedule); every && offsetDays != 0 {
		return nil, errors.New("an offset cannot be used with @every")
	}

	_, quartz := parsed.(*cronexpr.Expression)

	// robfig/cron keeps its own DST handling for the schedules it ran before.
	return cronSchedule{
		expr:         parsed,
		location:     loc,
		offsetDays:   offsetDays,
		useWallClock: quartz || offsetDays != 0,
	}, nil
}

// ValidateCronFormat reports whether a cron_format with the given offset can be scheduled.
func ValidateCronFormat(cronFormat string, offsetDays int) error {
	_, err := ParseCronSchedule(cronFormat, offsetDays, time.UTC)
	return err
}

// NextRuns returns up to n run times of s after t.
func NextRuns(s cron.Schedule, t time.Time, n int) []time.Time {
	runs := make([]time.Time, 0, n)

	for len(runs) < n {
		t = s.Next(t)
		if t.IsZero() {
			break
		}
		runs = append(runs, t)
	}

	return runs
}

// splitTimezone splits off a TZ= or CRON_TZ= prefix the way robfig/cron does,
// because robfig/cron panics on a prefix that has nothing after it.
func splitTimezone(cronFormat string) (*time.Location, string, error) {
	if !hasTimezonePrefix(cronFormat) {
		return nil, cronFormat, nil
	}

	prefix, expr, found := strings.Cut(cronFormat, " ")
	expr = strings.TrimSpace(expr)

	if !found || expr == "" {
		return nil, "", fmt.Errorf("missing schedule after %s", prefix)
	}

	if hasTimezonePrefix(expr) {
		return nil, "", errors.New("more than one time zone prefix")
	}

	_, name, _ := strings.Cut(prefix, "=")

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, "", fmt.Errorf("provided bad location %s: %w", name, err)
	}

	// robfig/cron evaluates a TZ=Local schedule in the scheduler's zone.
	if loc == time.Local {
		return nil, expr, nil
	}

	return loc, expr, nil
}

func hasTimezonePrefix(s string) bool {
	return strings.HasPrefix(s, "TZ=") || strings.HasPrefix(s, "CRON_TZ=")
}

// parseExpression keeps robfig/cron for every expression it accepts, so existing
// schedules run as before. Expressions with a Quartz day form go to cronexpr:
// #1 to #5 or L after a weekday, or L, LW or a day followed by W in the day of
// month.
func parseExpression(expr string) (cron.Schedule, error) {
	fields := strings.Fields(expr)

	if len(fields) != 5 || !usesQuartzDays(fields[2], fields[4]) {
		return cron.ParseStandard(expr)
	}

	dayOfMonth, dayOfWeek, err := quartzDays(fields[2], fields[4])
	if err != nil {
		return nil, err
	}

	// robfig/cron reads the other fields, so they mean what they mean in any other
	// schedule, and cronexpr gets plain lists of values.
	spec, err := cron.ParseStandard(fmt.Sprintf("%s %s * %s *", fields[0], fields[1], fields[3]))
	if err != nil {
		return nil, err
	}

	times, ok := spec.(*cron.SpecSchedule)
	if !ok {
		return nil, fmt.Errorf("unsupported schedule with L, W or #: %s", expr)
	}

	parsed, err := cronexpr.Parse(strings.Join([]string{
		values(times.Minute, 0, 59),
		values(times.Hour, 0, 23),
		dayOfMonth,
		values(times.Month, 1, 12),
		dayOfWeek,
	}, " "))
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

// usesQuartzDays checks only the day fields: month names such as JUL contain an L.
func usesQuartzDays(dayOfMonth, dayOfWeek string) bool {
	return strings.ContainsAny(dayOfMonth, "LWlw") || strings.ContainsAny(dayOfWeek, "#Ll")
}

// quartzDays checks the day fields of an expression with a Quartz day form.
// It accepts only the forms listed in parseExpression, because cronexpr
// misreads others, such as 02#2, and panics on some ranges.
func quartzDays(dayOfMonth, dayOfWeek string) (string, string, error) {
	// The libraries disagree on whether forms like */1 restrict a day field,
	// which decides between OR and AND. Quartz leaves one day field open.
	if !isOpenDay(dayOfMonth) && !isOpenDay(dayOfWeek) {
		return "", "", errors.New("day of month or day of week must be * or ? when the other uses L, W or #")
	}

	if !isOpenDay(dayOfMonth) {
		if !isQuartzMonthDay(dayOfMonth) {
			return "", "", fmt.Errorf("day of month %s must be L, LW or a day followed by W, such as 15W", dayOfMonth)
		}
		return dayOfMonth, "*", nil
	}

	for _, item := range strings.Split(dayOfWeek, ",") {
		if !isQuartzWeekday(item) {
			return "", "", fmt.Errorf("day of week %s must list weekdays followed by #1 to #5 or L, such as 2#2 or 5L", dayOfWeek)
		}
	}

	return "*", dayOfWeek, nil
}

func isOpenDay(field string) bool {
	return field == "*" || field == "?"
}

func isQuartzMonthDay(field string) bool {
	field = strings.ToUpper(field)

	if field == "L" || field == "LW" {
		return true
	}

	day, found := strings.CutSuffix(field, "W")
	n, err := strconv.Atoi(day)

	return found && err == nil && n >= 1 && n <= 31 && strconv.Itoa(n) == day
}

func isQuartzWeekday(item string) bool {
	item = strings.ToUpper(item)

	weekday, nth, found := strings.Cut(item, "#")
	if found {
		if len(nth) != 1 || nth[0] < '1' || nth[0] > '5' {
			return false
		}
	} else if weekday, found = strings.CutSuffix(item, "L"); !found {
		return false
	}

	switch weekday {
	case "0", "1", "2", "3", "4", "5", "6", "7", "SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT":
		return true
	}

	return false
}

// values lists the values set in a field of a robfig/cron schedule.
func values(field uint64, from, to uint) string {
	var list []string

	for v := from; v <= to; v++ {
		if field&(1<<v) != 0 {
			list = append(list, strconv.FormatUint(uint64(v), 10))
		}
	}

	return strings.Join(list, ",")
}
