package db

import (
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type AlertType string

const (
	AlertTypeEmail      AlertType = "email"
	AlertTypeTelegram   AlertType = "telegram"
	AlertTypeSlack      AlertType = "slack"
	AlertTypeTeams      AlertType = "teams"
	AlertTypeRocketChat AlertType = "rocketchat"
	AlertTypeDingTalk   AlertType = "dingtalk"
	AlertTypeGotify     AlertType = "gotify"
)

const (
	AlertModeInherit = "inherit"
	AlertModeDefault = "default"
	AlertModeIDs     = "ids"
)

// Alert is a named project notification: channel type + destination + message body.
type Alert struct {
	ID         int       `db:"id" json:"id" backup:"-"`
	ProjectID  int       `db:"project_id" json:"project_id" backup:"-"`
	Name       string    `db:"name" json:"name"`
	Type       AlertType `db:"type" json:"type"`
	Enabled    bool      `db:"enabled" json:"enabled"`
	IsDefault  bool      `db:"is_default" json:"is_default"`
	ChatID     *string   `db:"chat_id" json:"chat_id,omitempty"`
	ThreadID   *string   `db:"thread_id" json:"thread_id,omitempty"`
	URL        *string   `db:"url" json:"url,omitempty"`
	Token      *string   `db:"token" json:"token,omitempty"`
	Recipients *string   `db:"recipients" json:"recipients,omitempty"`
	KeyID      *int      `db:"key_id" json:"key_id,omitempty" backup:"-"`
	Body       *string   `db:"body" json:"body,omitempty"`
}

func (a *Alert) Validate() error {
	if a.Name == "" {
		return common_errors.NewValidationError("alert name can not be empty")
	}
	switch a.Type {
	case AlertTypeEmail, AlertTypeTelegram, AlertTypeSlack, AlertTypeTeams,
		AlertTypeRocketChat, AlertTypeDingTalk, AlertTypeGotify:
	default:
		return common_errors.NewValidationError("invalid alert type: " + string(a.Type))
	}
	if err := ValidateAlertURL(stringValue(a.URL)); err != nil {
		return err
	}
	switch a.Type {
	case AlertTypeTelegram:
		if stringValue(a.ChatID) == "" {
			return common_errors.NewValidationError("telegram chat id can not be empty")
		}
		if err := ValidateTelegramThreadID(stringValue(a.ThreadID)); err != nil {
			return err
		}
	case AlertTypeSlack, AlertTypeTeams, AlertTypeRocketChat, AlertTypeDingTalk:
		if stringValue(a.URL) == "" {
			return common_errors.NewValidationError("alert URL can not be empty")
		}
	case AlertTypeGotify:
		if stringValue(a.URL) != "" && stringValue(a.Token) == "" {
			return common_errors.NewValidationError("gotify token is missing for the alert URL")
		}
	}
	return nil
}

// Normalize turns blank optional strings into nil so the DB stores NULL
// instead of empty strings from the UI, and drops destination fields that
// do not apply to the selected type.
func (a *Alert) Normalize() {
	a.Name = strings.TrimSpace(a.Name)
	a.ChatID = emptyToNil(a.ChatID)
	a.ThreadID = emptyToNil(a.ThreadID)
	a.URL = emptyToNil(a.URL)
	a.Token = emptyToNil(a.Token)
	a.Recipients = emptyToNil(a.Recipients)
	a.Body = emptyToNil(a.Body)

	switch a.Type {
	case AlertTypeTelegram:
		a.URL = nil
		a.Token = nil
		a.Recipients = nil
	case AlertTypeEmail:
		a.ChatID = nil
		a.ThreadID = nil
		a.URL = nil
		a.Token = nil
	case AlertTypeSlack, AlertTypeTeams, AlertTypeRocketChat, AlertTypeDingTalk:
		a.ChatID = nil
		a.ThreadID = nil
		a.Token = nil
		a.Recipients = nil
	case AlertTypeGotify:
		a.ChatID = nil
		a.ThreadID = nil
		a.Recipients = nil
	}
}

// MarshalJSON omits token so list/get/create responses never echo the secret.
// Backup uses the backup/db tags via reflection, not this method.
func (a Alert) MarshalJSON() ([]byte, error) {
	type alertJSON Alert
	out := alertJSON(a)
	out.Token = nil
	return json.Marshal(out)
}

// ValidateAlertURL allows empty (Gotify instance pair) and http(s) destinations.
// Loopback, unspecified, and link-local hosts are rejected to limit SSRF
// from project webhooks. Private RFC1918 addresses stay allowed so a
// self-hosted Gotify or Rocket.Chat still works.
func ValidateAlertURL(raw string) error {
	return validateAlertURL(raw, false)
}

func ValidateGotifyURL(raw string) error {
	return ValidateAlertURL(raw)
}

func ValidateTelegramThreadID(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if _, err := strconv.Atoi(raw); err != nil {
		return common_errors.NewValidationError("telegram thread id must be numeric")
	}
	return nil
}

func validateAlertURL(raw string, requireHTTPS bool) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return common_errors.NewValidationError("invalid alert URL")
	}
	scheme := strings.ToLower(u.Scheme)
	if requireHTTPS {
		if scheme != "https" {
			return common_errors.NewValidationError("gotify URL must be https")
		}
	} else if scheme != "http" && scheme != "https" {
		return common_errors.NewValidationError("alert URL must be http or https")
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "metadata" || host == "metadata.google.internal" ||
		strings.HasSuffix(host, ".metadata.google.internal") {
		return common_errors.NewValidationError("alert URL host is not allowed")
	}
	if ip := net.ParseIP(host); ip != nil && AlertIPForbidden(ip) {
		return common_errors.NewValidationError("alert URL host is not allowed")
	}
	return nil
}

// AlertIPForbidden reports whether a resolved address must not be dialed for
// user-controlled alert webhooks.
func AlertIPForbidden(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

// AllowedAlertIPs returns the subset of addresses that alertHTTPClient may dial.
func AllowedAlertIPs(ips []net.IP) []net.IP {
	var out []net.IP
	for _, ip := range ips {
		if !AlertIPForbidden(ip) {
			out = append(out, ip)
		}
	}
	return out
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func emptyToNil(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	return &trimmed
}

func (a Alert) GetID() int {
	return a.ID
}

func (a Alert) GetName() string {
	return a.Name
}

// ResolveAlertOn prefers the new alert_on_* field. If it is omitted, the
// legacy suppress_* flag is inverted so older API clients keep working.
func ResolveAlertOn(on *bool, suppress bool) bool {
	if on != nil {
		return *on
	}
	return !suppress
}

func BoolPtr(v bool) *bool {
	return &v
}
