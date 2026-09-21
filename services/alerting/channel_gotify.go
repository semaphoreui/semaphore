package alerting

import (
	"context"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/util"
)

const TypeGotify db.AlertType = "gotify"

// gotifyChannel pushes to a Gotify server. A project alert may point at its
// own server (url + token); leaving both empty reuses the server-wide pair
// from config.json.
type gotifyChannel struct {
	channelBase
}

func newGotifyChannel() Channel {
	return &gotifyChannel{
		channelBase: channelBase{
			typ:   TypeGotify,
			title: "Gotify",
			icon:  "mdi-bell-ring-outline",
			fields: []Field{
				{Name: FieldURL},
				{Name: FieldToken, Secret: true},
			},
			defaultEvents: chatEvents(),
			format:        BodyFormatText,
			templateFile:  "gotify.tmpl",
		},
	}
}

func (c *gotifyChannel) Validate(dest Destination) error {
	url := strings.TrimSpace(dest.URL)
	token := strings.TrimSpace(dest.Token)
	if url == "" && token == "" {
		return nil // server-wide pair
	}
	if url == "" || token == "" {
		return common_errors.NewValidationError("gotify needs both URL and token, or neither to use the server settings")
	}
	return ValidateWebhookURL(url)
}

func (c *gotifyChannel) InstanceDestination(cfg *util.ConfigType, _ db.Project) (Destination, bool) {
	if cfg == nil || !cfg.GotifyAlert || strings.TrimSpace(cfg.GotifyUrl) == "" || strings.TrimSpace(cfg.GotifyToken) == "" {
		return Destination{}, false
	}
	return Destination{
		URL:     strings.TrimSpace(cfg.GotifyUrl),
		Token:   strings.TrimSpace(cfg.GotifyToken),
		Trusted: true,
	}, true
}

func (c *gotifyChannel) Send(ctx context.Context, cfg *util.ConfigType, dest Destination, msg Message) error {
	target, token, trusted, err := c.resolve(cfg, dest)
	if err != nil {
		return err
	}
	return postJSON(
		ctx,
		Destination{Trusted: trusted},
		strings.TrimRight(target, "/")+"/message",
		msg.Body,
		map[string]string{"X-Gotify-Key": token},
		200,
	)
}

// resolve picks the alert's own server or falls back to config.json. The
// fallback pair is admin supplied and therefore trusted.
func (c *gotifyChannel) resolve(cfg *util.ConfigType, dest Destination) (url string, token string, trusted bool, err error) {
	url = strings.TrimSpace(dest.URL)
	token = strings.TrimSpace(dest.Token)
	if url != "" && token != "" {
		return url, token, dest.Trusted, nil
	}
	if url != "" || token != "" {
		return "", "", false, common_errors.NewValidationError("gotify needs both URL and token, or neither to use the server settings")
	}
	if cfg == nil || strings.TrimSpace(cfg.GotifyUrl) == "" || strings.TrimSpace(cfg.GotifyToken) == "" {
		return "", "", false, common_errors.NewValidationError("gotify is not configured on the server (gotify_url, gotify_token)")
	}
	return strings.TrimSpace(cfg.GotifyUrl), strings.TrimSpace(cfg.GotifyToken), true, nil
}
