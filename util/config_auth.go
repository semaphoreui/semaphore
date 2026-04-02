package util

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

type JWTAuthConfig struct {
	Enabled       bool   `json:"enabled" env:"SEMAPHORE_JWT_AUTH_ENABLED"`
	Header        string `json:"header" env:"SEMAPHORE_JWT_AUTH_HEADER"`
	JWKSURL       string `json:"jwks_url" env:"SEMAPHORE_JWT_AUTH_JWKS_URL"`
	Audience      string `json:"audience" env:"SEMAPHORE_JWT_AUTH_AUDIENCE"`
	Issuer        string `json:"issuer" env:"SEMAPHORE_JWT_AUTH_ISSUER"`
	UsernameClaim string `json:"username_claim" env:"SEMAPHORE_JWT_AUTH_USERNAME_CLAIM"`
	NameClaim     string `json:"name_claim" env:"SEMAPHORE_JWT_AUTH_NAME_CLAIM"`
	EmailClaim    string `json:"email_claim" env:"SEMAPHORE_JWT_AUTH_EMAIL_CLAIM"`
}

func (j *JWTAuthConfig) GetUsernameClaim() string {
	if j.UsernameClaim == "" {
		return "email"
	}
	return j.UsernameClaim
}

func (j *JWTAuthConfig) GetEmailClaim() string {
	if j.EmailClaim == "" {
		return "email"
	}
	return j.EmailClaim
}

func (j *JWTAuthConfig) GetNameClaim() string {
	if j.NameClaim == "" {
		return "name"
	}
	return j.NameClaim
}

func (j *JWTAuthConfig) GetHeader() string {
	return j.Header
}

type AuthConfig struct {
	Totp  *TotpConfig      `json:"totp,omitempty"`
	Email *EmailAuthConfig `json:"email,omitempty"`
	JWT   *JWTAuthConfig   `json:"jwt,omitempty"`
}
