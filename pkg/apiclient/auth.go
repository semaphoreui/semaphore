package apiclient

import (
	"context"
	"net/http"
	"time"

	"github.com/semaphoreui/semaphore/db"
)

// Login opens a session with a login (or e-mail) and a password. The session
// cookie stays in the client's in-memory jar; the server answers 204 whether
// or not a second factor is still pending, so call Me afterwards: a 401 with
// TOTP enabled on the server means Verify is required.
func (c *Client) Login(ctx context.Context, auth, password string) error {
	body := struct {
		Auth     string `json:"auth"`
		Password string `json:"password"`
	}{Auth: auth, Password: password}
	return c.Do(ctx, http.MethodPost, "/auth/login", body, nil)
}

// Verify completes a session that waits for a TOTP passcode.
func (c *Client) Verify(ctx context.Context, passcode string) error {
	body := struct {
		Passcode string `json:"passcode"`
	}{Passcode: passcode}
	return c.Do(ctx, http.MethodPost, "/auth/verify", body, nil)
}

// Logout expires the session held in the cookie jar. It is a no-op for a
// token-only client.
func (c *Client) Logout(ctx context.Context) error {
	return c.Do(ctx, http.MethodPost, "/auth/logout", nil, nil)
}

// CreateToken issues a new API token for the authenticated user. The full
// token value is returned only here: ListTokens truncates ids to 8 characters.
func (c *Client) CreateToken(ctx context.Context, name string, expiresAt *time.Time) (db.APIToken, error) {
	body := struct {
		Name      string     `json:"name,omitempty"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
	}{Name: name, ExpiresAt: expiresAt}
	var token db.APIToken
	err := c.Do(ctx, http.MethodPost, "/user/tokens", body, &token)
	return token, err
}

// ListTokens returns the user's API tokens with ids truncated to 8 characters.
func (c *Client) ListTokens(ctx context.Context) ([]db.APIToken, error) {
	var tokens []db.APIToken
	err := c.Do(ctx, http.MethodGet, "/user/tokens", nil, &tokens)
	return tokens, err
}

// DeleteToken revokes a token by its id or 8-character prefix as listed.
func (c *Client) DeleteToken(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/user/tokens/"+id, nil, nil)
}
