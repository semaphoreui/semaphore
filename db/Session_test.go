package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession_IsExpiredAt(t *testing.T) {
	now := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	maxLife := 24 * time.Hour

	tests := []struct {
		name     string
		session  Session
		maxLife  time.Duration
		expected bool
	}{
		{
			name:     "fresh session is valid",
			session:  Session{Created: now.Add(-time.Hour), LastActive: now.Add(-time.Minute)},
			maxLife:  maxLife,
			expected: false,
		},
		{
			name:     "revoked session is expired",
			session:  Session{Created: now, LastActive: now, Expired: true},
			maxLife:  maxLife,
			expected: true,
		},
		{
			name:     "unused longer than inactivity timeout is expired",
			session:  Session{Created: now.Add(-8 * 24 * time.Hour), LastActive: now.Add(-SessionInactivityTimeout - time.Second)},
			maxLife:  0,
			expected: true,
		},
		{
			name:     "unused exactly inactivity timeout is still valid",
			session:  Session{Created: now.Add(-8 * 24 * time.Hour), LastActive: now.Add(-SessionInactivityTimeout)},
			maxLife:  0,
			expected: false,
		},
		{
			name:     "older than max life is expired even when recently active",
			session:  Session{Created: now.Add(-maxLife - time.Second), LastActive: now},
			maxLife:  maxLife,
			expected: true,
		},
		{
			name:     "exactly max life old is still valid",
			session:  Session{Created: now.Add(-maxLife), LastActive: now},
			maxLife:  maxLife,
			expected: false,
		},
		{
			name:     "zero max life disables absolute limit",
			session:  Session{Created: now.Add(-365 * 24 * time.Hour), LastActive: now},
			maxLife:  0,
			expected: false,
		},
		{
			name:     "negative max life disables absolute limit",
			session:  Session{Created: now.Add(-365 * 24 * time.Hour), LastActive: now},
			maxLife:  -time.Hour,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.IsExpiredAt(now, tt.maxLife, SessionInactivityTimeout))
		})
	}
}
