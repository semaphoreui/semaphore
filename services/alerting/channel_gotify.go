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
				{Name: FieldURL, Kind: FieldKindText, Override: true},
			},
			defaultEvents: chatEvents(),
			format:        BodyFormatText,
			templateFile:  "gotify.tmpl",
		},
	}
}

func (c *gotifyChannel) Secret() *SecretSpec {
	return &SecretSpec{KeyType: db.AccessKeyString, Label: "Application token"}
}

func (c *gotifyChannel) Validate(cfg *util.ConfigType, dest Destination) error {
	_, _, _, err := c.resolve(cfg, dest)
	return err
}

func (c *gotifyChannel) ServerReady(cfg *util.ConfigType) error {
	if cfg == nil || strings.TrimSpace(cfg.GotifyUrl) == "" || strings.TrimSpace(cfg.GotifyToken) == "" {
		return common_errors.NewValidationError("gotify is not configured on the server (gotify_url, gotify_token)")
	}
	return nil
}

func (c *gotifyChannel) InstanceDestination(cfg *util.ConfigType, _ db.Project) (Destination, bool) {
	if cfg == nil || !cfg.GotifyAlert || strings.TrimSpace(cfg.GotifyUrl) == "" || strings.TrimSpace(cfg.GotifyToken) == "" {
		return Destination{}, false
	}
	return Destination{
		URL:     strings.TrimSpace(cfg.GotifyUrl),
		Secret:  &db.AccessKey{Type: db.AccessKeyString, String: strings.TrimSpace(cfg.GotifyToken)},
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

// resolve picks the alert's own server (URL + access key) or falls back to
// the config.json pair, which is admin supplied and therefore trusted.
func (c *gotifyChannel) resolve(cfg *util.ConfigType, dest Destination) (url string, token string, trusted bool, err error) {
	url = strings.TrimSpace(dest.URL)
	if dest.Trusted && url != "" {
		return url, strings.TrimSpace(dest.Secret.String), true, nil
	}
	if url == "" && dest.Secret == nil {
		if err = c.ServerReady(cfg); err != nil {
			return "", "", false, err
		}
		return strings.TrimSpace(cfg.GotifyUrl), strings.TrimSpace(cfg.GotifyToken), true, nil
	}
	if url == "" || dest.Secret == nil {
		return "", "", false, common_errors.NewValidationError("gotify needs both a server URL and an access key with the application token, or neither to use the server settings")
	}
	if err = ValidateWebhookURL(url); err != nil {
		return "", "", false, err
	}
	token = strings.TrimSpace(dest.Secret.String)
	if token == "" {
		return "", "", false, common_errors.NewValidationError("the selected access key holds no gotify token")
	}
	return url, token, false, nil
}
