//go:build !pro

package util

type AuthConfig struct {
	Totp *TotpConfig `json:"totp,omitempty"`
}
