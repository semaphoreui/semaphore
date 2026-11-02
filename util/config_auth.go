package util

import "time"

// DefaultSessionInactivityTimeout is how long a session may stay unused before
// it is rejected when auth.session_inactivity_hours is not configured.
const DefaultSessionInactivityTimeout = 7 * 24 * time.Hour

// AuthConfig holds settings that apply to every login method.
type AuthConfig struct {
	// SessionInactivityHours is how long a login session may stay unused, in
	// hours, before it is rejected and expired. Every authenticated request
	// resets the counter. 0 (default) means the built-in 168 hours (7 days).
	// The inactivity timeout cannot be disabled.
	SessionInactivityHours int `json:"session_inactivity_hours,omitempty" rule:"^[0-9]*$" env:"SEMAPHORE_SESSION_INACTIVITY_HOURS"`
}

// SessionInactivityTimeout returns how long a session may stay unused before
// it is rejected, as configured by auth.session_inactivity_hours, or
// DefaultSessionInactivityTimeout when the option is not set. It is safe to
// call when the auth section is not configured.
func (c *ConfigType) SessionInactivityTimeout() time.Duration {
	if c == nil || c.Auth == nil || c.Auth.SessionInactivityHours <= 0 {
		return DefaultSessionInactivityTimeout
	}
	return time.Duration(c.Auth.SessionInactivityHours) * time.Hour
}

type RecaptchaConfig struct {
	Enabled string `json:"enabled,omitempty" env:"SEMAPHORE_RECAPTCHA_ENABLED"`
	SiteKey string `json:"site_key,omitempty" env:"SEMAPHORE_RECAPTCHA_SITE_KEY"`
}

type EmailAuthConfig struct {
	Enabled                  bool     `json:"enabled" env:"SEMAPHORE_EMAIL_2TP_ENABLED"`
	AllowLoginAsExternalUser bool     `json:"allow_login_as_external_user" env:"SEMAPHORE_EMAIL_2TP_ALLOW_LOGIN_AS_EXTERNAL_USER"`
	AllowCreateExternalUsers bool     `json:"allow_create_external_user" env:"SEMAPHORE_EMAIL_2TP_ALLOW_CREATE_EXTERNAL_USER"`
	AllowedDomains           []string `json:"allowed_domains" env:"SEMAPHORE_EMAIL_2TP_ALLOWED_DOMAINS"`
	DisableForOidc           bool     `json:"disable_for_oidc" env:"SEMAPHORE_EMAIL_2TP_DISABLE_FOR_OIDC"`
}

type MultifactorAuthConfig struct {
	Totp  *TotpConfig      `json:"totp,omitempty"`
	Email *EmailAuthConfig `json:"email,omitempty"`
}
