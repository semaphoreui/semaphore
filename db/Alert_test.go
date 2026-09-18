package db

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAlertOn(t *testing.T) {
	assert.True(t, ResolveAlertOn(nil, false))
	assert.False(t, ResolveAlertOn(nil, true))
	assert.True(t, ResolveAlertOn(BoolPtr(true), true))
	assert.False(t, ResolveAlertOn(BoolPtr(false), false))
}

func TestAlertValidate(t *testing.T) {
	alert := Alert{Name: "", Type: AlertTypeTelegram}
	assert.Error(t, alert.Validate())

	alert.Name = "ops"
	alert.Type = "discord"
	assert.Error(t, alert.Validate())

	alert.Type = AlertTypeTelegram
	assert.Error(t, alert.Validate())
	chatID := "12345"
	alert.ChatID = &chatID
	assert.NoError(t, alert.Validate())

	bad := "javascript:alert(1)"
	alert.URL = &bad
	assert.Error(t, alert.Validate())

	meta := "http://169.254.169.254/latest/meta-data"
	alert.URL = &meta
	assert.Error(t, alert.Validate())

	okURL := "https://hooks.slack.com/services/T/B/X"
	alert.URL = &okURL
	assert.NoError(t, alert.Validate())

	httpGotify := "http://gotify.example.com"
	gotify := Alert{Name: "gotify", Type: AlertTypeGotify, URL: &httpGotify}
	assert.Error(t, gotify.Validate())

	httpsGotify := "https://gotify.example.com"
	gotify.URL = &httpsGotify
	assert.Error(t, gotify.Validate())

	token := "secret"
	gotify.Token = &token
	assert.NoError(t, gotify.Validate())

	gotify.URL = nil
	gotify.Token = nil
	assert.NoError(t, gotify.Validate())

	gotify.URL = &httpGotify
	gotify.Token = &token
	assert.NoError(t, gotify.Validate())

	badThread := "general"
	tg := Alert{Name: "tg", Type: AlertTypeTelegram, ChatID: &chatID, ThreadID: &badThread}
	assert.Error(t, tg.Validate())
	okThread := "12"
	tg.ThreadID = &okThread
	assert.NoError(t, tg.Validate())
}

func TestValidateAlertURL(t *testing.T) {
	assert.NoError(t, ValidateAlertURL(""))
	assert.NoError(t, ValidateAlertURL("https://example.com/hook"))
	assert.NoError(t, ValidateAlertURL("http://10.0.0.8/hooks"))
	assert.Error(t, ValidateAlertURL("ftp://example.com/hook"))
	assert.Error(t, ValidateAlertURL("http://metadata.google.internal/"))
	assert.Error(t, ValidateAlertURL("http://localhost/hook"))
	assert.Error(t, ValidateAlertURL("http://127.0.0.1/hook"))
	assert.Error(t, ValidateAlertURL("http://[::1]/hook"))
	assert.Error(t, ValidateAlertURL("http://169.254.169.254/latest/meta-data"))
	assert.NoError(t, ValidateGotifyURL("http://gotify.example.com"))
	assert.NoError(t, ValidateGotifyURL("https://gotify.example.com"))
	assert.NoError(t, ValidateGotifyURL(""))
}

func TestAllowedAlertIPs(t *testing.T) {
	assert.True(t, AlertIPForbidden(net.ParseIP("127.0.0.1")))
	assert.True(t, AlertIPForbidden(net.ParseIP("::1")))
	assert.True(t, AlertIPForbidden(net.ParseIP("169.254.169.254")))
	assert.True(t, AlertIPForbidden(net.ParseIP("0.0.0.0")))
	assert.False(t, AlertIPForbidden(net.ParseIP("10.0.0.8")))
	assert.False(t, AlertIPForbidden(net.ParseIP("1.1.1.1")))

	allowed := AllowedAlertIPs([]net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("10.0.0.8"),
		net.ParseIP("169.254.169.254"),
	})
	require.Len(t, allowed, 1)
	assert.Equal(t, "10.0.0.8", allowed[0].String())
}

func TestAlert_TokenOmittedFromJSON(t *testing.T) {
	token := "secret-token"
	raw, err := json.Marshal(Alert{
		ID:      3,
		Name:    "Gotify",
		Type:    AlertTypeGotify,
		Enabled: true,
		Token:   &token,
	})
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "secret-token")
	assert.NotContains(t, string(raw), `"token"`)
}

func TestAlertNormalize(t *testing.T) {
	blank := "  "
	url := "https://hooks.slack.com/x"
	alert := Alert{
		Name:   "  Ops  ",
		Type:   AlertTypeSlack,
		ChatID: &blank,
		Body:   &blank,
		URL:    &url,
		Token:  &blank,
	}
	alert.Normalize()
	assert.Equal(t, "Ops", alert.Name)
	assert.Nil(t, alert.ChatID)
	assert.Nil(t, alert.Token)
	assert.Nil(t, alert.Recipients)
	assert.Equal(t, "https://hooks.slack.com/x", *alert.URL)
}
