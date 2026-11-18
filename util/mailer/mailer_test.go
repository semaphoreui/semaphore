package mailer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/semaphoreui/semaphore/pkg/tz"
)

func TestFormatMailDate_UsesNumericTimezoneOffset(t *testing.T) {
	formatted := formatMailDate(tz.Now())

	assert.Regexp(t, `[+-]\d{4}$`, formatted)
	assert.NotContains(t, formatted, "UTC")
}

func TestFormatMailDate_FixedInstant(t *testing.T) {
	instant := time.Date(2026, 9, 21, 20, 0, 0, 0, time.UTC)

	formatted := formatMailDate(instant)

	assert.Equal(t, "Mon, 21 Sep 2026 20:00:00 +0000", formatted)
}
