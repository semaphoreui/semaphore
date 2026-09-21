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
}

func newTelegramChannel() Channel {
	return &telegramChannel{
		channelBase: channelBase{
			typ:   TypeTelegram,
			title: "Telegram",
			icon:  "mdi-send-circle-outline",
			fields: []Field{
				{Name: FieldChatID, Required: true},
				{Name: FieldThreadID},
			},
			defaultEvents: chatEvents(),
			format:        BodyFormatText,
			templateFile:  "telegram.tmpl",
		},
	}
}

func (c *telegramChannel) Validate(dest Destination) error {
	if strings.TrimSpace(dest.ChatID) == "" {
		return common_errors.NewValidationError("telegram chat id can not be empty")
	}
	return validateThreadID(dest.ThreadID)
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

func (c *telegramChannel) ServerReady(cfg *util.ConfigType) error {
	if cfg == nil || !cfg.TelegramAlert || strings.TrimSpace(cfg.TelegramToken) == "" {
		return common_errors.NewValidationError("telegram bot token is not configured on the server (telegram_alert, telegram_token)")
	}
	return nil
}

func (c *telegramChannel) Send(ctx context.Context, cfg *util.ConfigType, dest Destination, msg Message) error {
	if err := c.ServerReady(cfg); err != nil {
		return err
	}
	if err := c.Validate(dest); err != nil {
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
	return postJSON(ctx, apiDest, telegramAPI+cfg.TelegramToken+"/sendMessage", string(body), nil, 200)
}
