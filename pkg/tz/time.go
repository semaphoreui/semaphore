package tz

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func In(t time.Time) time.Time {
	return t.In(time.UTC)
}
