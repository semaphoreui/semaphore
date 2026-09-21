package db

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAlertEvents(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected AlertEvents
	}{
		{"empty", "", nil},
		{"blank", "  ", nil},
		{"single", "error", AlertEvents{AlertEventError}},
		{"spaces and duplicates", " success, error ,success", AlertEvents{AlertEventSuccess, AlertEventError}},
		{"trailing comma", "waiting_confirmation,", AlertEvents{AlertEventWaitingConfirmation}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseAlertEvents(tt.raw))
		})
	}
}

func TestAlertEvents_ScanValueRoundTrip(t *testing.T) {
	events := AlertEvents{AlertEventSuccess, AlertEventError}

	value, err := events.Value()
	require.NoError(t, err)
	assert.Equal(t, "success,error", value)

	var scanned AlertEvents
	require.NoError(t, scanned.Scan([]byte("success,error")))
	assert.Equal(t, events, scanned)

	require.NoError(t, scanned.Scan(nil))
	assert.Nil(t, scanned)

	emptyValue, err := AlertEvents{}.Value()
	require.NoError(t, err)
	assert.Nil(t, emptyValue)

	assert.Error(t, scanned.Scan(42))
}

func TestAlertEvents_Contains(t *testing.T) {
	events := AlertEvents{AlertEventError}
	assert.True(t, events.Contains(AlertEventError))
	assert.False(t, events.Contains(AlertEventSuccess))
	assert.False(t, AlertEvents(nil).Contains(AlertEventError))
}

func TestAlert_Normalize(t *testing.T) {
	blank := "   "
	url := " https://hooks.example/x "
	alert := Alert{
		Name:   "  Ops  ",
		Type:   " Slack ",
		ChatID: &blank,
		URL:    &url,
	}

	alert.Normalize()

	assert.Equal(t, "Ops", alert.Name)
	assert.Equal(t, AlertType("slack"), alert.Type)
	assert.Nil(t, alert.ChatID)
	require.NotNil(t, alert.URL)
	assert.Equal(t, "https://hooks.example/x", *alert.URL)
}

func TestAlert_Validate(t *testing.T) {
	tests := []struct {
		name    string
		alert   Alert
		wantErr string
	}{
		{"valid", Alert{Name: "a", Type: "slack"}, ""},
		{"missing name", Alert{Type: "slack"}, "name"},
		{"missing type", Alert{Name: "a"}, "type"},
		{"unknown event", Alert{Name: "a", Type: "slack", Events: AlertEvents{"started"}}, "unknown alert event"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.alert.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestAlert_MarshalJSON_HidesToken(t *testing.T) {
	token := "secret"
	alert := Alert{ID: 1, Name: "Gotify", Type: "gotify", Token: &token}

	out, err := json.Marshal(alert)
	require.NoError(t, err)

	assert.NotContains(t, string(out), "secret")
	assert.Contains(t, string(out), `"events":[]`)
}

func TestAlertSnapshot_Allows(t *testing.T) {
	snap := AlertSnapshot{OnSuccess: false, OnError: true}
	assert.False(t, snap.Allows(AlertEventSuccess))
	assert.True(t, snap.Allows(AlertEventError))
	assert.True(t, snap.Allows(AlertEventWaitingConfirmation))
}

func TestAlertSnapshot_ScanValueRoundTrip(t *testing.T) {
	snap := AlertSnapshot{Instance: true, AlertIDs: []int{3, 4}, OnSuccess: true, OnError: false}

	value, err := snap.Value()
	require.NoError(t, err)

	var scanned AlertSnapshot
	require.NoError(t, scanned.Scan(value))
	assert.Equal(t, snap, scanned)

	require.NoError(t, scanned.Scan(nil))
	assert.Equal(t, AlertSnapshot{}, scanned)
	assert.True(t, scanned.IsEmpty())
}

func TestTemplate_NormalizeAlerts(t *testing.T) {
	tpl := Template{}
	require.NoError(t, tpl.NormalizeAlerts())
	assert.Equal(t, AlertModeDefault, tpl.AlertMode)

	tpl = Template{AlertIDs: []int{1}}
	require.NoError(t, tpl.NormalizeAlerts())
	assert.Equal(t, AlertModeIDs, tpl.AlertMode)

	tpl = Template{AlertMode: "inherit"}
	assert.Error(t, tpl.NormalizeAlerts())
}

func TestSchedule_NormalizeAlerts(t *testing.T) {
	s := Schedule{}
	require.NoError(t, s.NormalizeAlerts())
	assert.Equal(t, AlertModeInherit, s.AlertMode)

	s = Schedule{AlertMode: "default"}
	assert.Error(t, s.NormalizeAlerts())
}
