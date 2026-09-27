package alerting

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/util"
)

const TypeTelegram db.AlertType = "telegram"

const telegramAPI = "https://api.telegram.org/bot"

// telegramChannel posts through the Bot API. The bot token is a server-wide
// secret (telegram_token in config.json); project alerts only choose the
// chat and, optionally, the forum topic.
type telegramChannel struct {
	channelBase
	// api is the Bot API base; tests point it at a local server.
	api string
}

func newTelegramChannel() Channel {
	return &telegramChannel{
		channelBase: channelBase{
			typ:   TypeTelegram,
			title: "Telegram",
			icon:  "mdi-send-circle-outline",
			fields: []Field{
				{Name: FieldChatID, Kind: FieldKindText, Required: true},
				{Name: FieldThreadID, Kind: FieldKindText},
			},
			defaultEvents: chatEvents(),
			format:        BodyFormatHTML,
			templateFile:  "telegram.tmpl",
		},
	}
}

func (c *telegramChannel) Secret() *SecretSpec {
	return &SecretSpec{KeyType: db.AccessKeyString, Label: "Bot token"}
}

func (c *telegramChannel) Validate(cfg *util.ConfigType, dest Destination) error {
	if strings.TrimSpace(dest.ChatID) == "" {
		return common_errors.NewValidationError("telegram chat id can not be empty")
	}
	if err := validateThreadID(dest.ThreadID); err != nil {
		return err
	}
	if _, err := c.token(cfg, dest); err != nil {
		return err
	}
	return nil
}

// token is the alert's own bot token or, without a key, the server-wide one.
func (c *telegramChannel) token(cfg *util.ConfigType, dest Destination) (string, error) {
	if dest.Secret != nil {
		if tok := strings.TrimSpace(dest.Secret.String); tok != "" {
			return tok, nil
		}
		return "", common_errors.NewValidationError("the selected access key holds no telegram bot token")
	}
	if cfg == nil || strings.TrimSpace(cfg.TelegramToken) == "" {
		return "", common_errors.NewValidationError("telegram bot token is not configured on the server (telegram_token); select an access key with your own token")
	}
	return strings.TrimSpace(cfg.TelegramToken), nil
}

func validateThreadID(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		return common_errors.NewValidationError("telegram thread id must be a positive integer")
	}
	return nil
}

// InstanceDestination keeps the legacy behaviour: the project chat overrides
// the server-wide chat, and no chat at all means nothing is sent.
func (c *telegramChannel) InstanceDestination(cfg *util.ConfigType, project db.Project) (Destination, bool) {
	if cfg == nil || !cfg.TelegramAlert || cfg.TelegramToken == "" {
		return Destination{}, false
	}
	chat := strings.TrimSpace(cfg.TelegramChat)
	if project.AlertChat != nil && strings.TrimSpace(*project.AlertChat) != "" {
		chat = strings.TrimSpace(*project.AlertChat)
	}
	if chat == "" {
		return Destination{}, false
	}
	return Destination{ChatID: chat, Trusted: true}, true
}

func (c *telegramChannel) ServerEnabled(cfg *util.ConfigType) bool {
	return cfg != nil && cfg.TelegramAlert
}

func (c *telegramChannel) ServerReady(cfg *util.ConfigType) error {
	if cfg == nil || strings.TrimSpace(cfg.TelegramToken) == "" {
		return common_errors.NewValidationError("telegram bot token is not configured on the server (telegram_token)")
	}
	return nil
}

func (c *telegramChannel) apiBase() string {
	if c.api != "" {
		return c.api
	}
	return telegramAPI
}

func (c *telegramChannel) Send(ctx context.Context, cfg *util.ConfigType, dest Destination, msg Message) error {
	if err := c.Validate(cfg, dest); err != nil {
		return err
	}
	token, err := c.token(cfg, dest)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"chat_id":    dest.ChatID,
		"parse_mode": "HTML",
		"text":       msg.Body,
	}
	if thread := strings.TrimSpace(dest.ThreadID); thread != "" {
		n, _ := strconv.Atoi(thread)
		payload["message_thread_id"] = n
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// The Bot API host is fixed and never user supplied, so the request is
	// sent as trusted regardless of where the destination came from.
	apiDest := Destination{Trusted: true}
	return postJSON(ctx, apiDest, c.apiBase()+token+"/sendMessage", string(body), nil, 200)
}
