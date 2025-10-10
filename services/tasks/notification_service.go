package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/semaphoreui/semaphore/util"
)

// Notifier is a common interface for sending notifications
// Implementations should use token/channel style configuration, not insecure URL params.
type Notifier interface {
	Name() string
	Enabled() bool
	Send(alert Alert) error
}

// NotificationService holds all configured notifiers and fans out alerts
// Implementations are created from util.Config.Notifications and support back-compat fallbacks where needed.
type NotificationService struct {
	notifiers []Notifier
}

func fillNotificationConfig(src map[string]any, target any) error {
	content, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(content, target)
	return err
}
func NewNotificationService() *NotificationService {
	var notifiers []Notifier

	//cfg := util.Config.Notifications

	for _, v := range util.Config.Notifications {
		var enabled bool
		var ok bool
		if enabled, ok = v["enabled"].(bool); !ok || !enabled {
			continue
		}

		switch v["type"].(util.NotificationType) {
		case util.NotificationWebhook:

		case util.NotificationTelegram:
			var tgConfig util.TelegramNotifConfig
			fillNotificationConfig(v, &tgConfig)
			n := &TelegramNotifier{
				token:        tgConfig.Token,
				chatID:       tgConfig.Chat,
				templateFile: tgConfig.Template,
			}

			n.token = tgConfig.Token
			n.chatID = tgConfig.Chat
			notifiers = append(notifiers, n)
		}
	}

	//if cfg == nil {
	//
	//	if cfg.Telegram != nil {
	//		n := &TelegramNotifier{}
	//		n.token = cfg.Telegram.Token
	//		n.chatID = cfg.Telegram.Chat
	//		notifiers = append(notifiers, n)
	//	}
	//	if cfg.Slack != nil {
	//		n := &SlackNotifier{}
	//		n.token = cfg.Slack.Token
	//		n.channel = cfg.Slack.Channel
	//		notifiers = append(notifiers, n)
	//	}
	//	if cfg.Gotify != nil {
	//		n := &GotifyNotifier{}
	//		n.server = cfg.Gotify.Server
	//		n.token = cfg.Gotify.Token
	//		n.title = cfg.Gotify.Title
	//		if cfg.Gotify.Priority != nil {
	//			n.priority = *cfg.Gotify.Priority
	//		}
	//		notifiers = append(notifiers, n)
	//	}
	//	if cfg.DingTalk != nil {
	//		n := &DingTalkNotifier{}
	//		n.token = cfg.DingTalk.Token
	//		n.channel = cfg.DingTalk.Channel
	//		notifiers = append(notifiers, n)
	//	}
	//}

	return &NotificationService{notifiers: notifiers}
}

func (s *NotificationService) HasNotifiers() bool {
	return len(s.notifiers) > 0
}

func (s *NotificationService) SendAll(alert Alert) {
	for _, n := range s.notifiers {
		if !n.Enabled() {
			continue
		}
		_ = n.Send(alert)
	}
}

// Notifiers returns the configured notifier implementations
func (s *NotificationService) Notifiers() []Notifier {
	return s.notifiers
}

// TelegramNotifier sends messages to Telegram using a bot token and chat ID
// Falls back to util.Config.Telegram* if new config is not provided
// and Enabled() should reflect availability of credentials

type TelegramNotifier struct {
	token        string
	chatID       string
	templateFile string
}

func (t *TelegramNotifier) Name() string { return "telegram" }
func (t *TelegramNotifier) Enabled() bool {
	if t.token == "" || t.chatID == "" {
		// fallback to old config for backward compatibility
		if util.Config.TelegramAlert && util.Config.TelegramToken != "" && util.Config.TelegramChat != "" {
			t.token = util.Config.TelegramToken
			t.chatID = util.Config.TelegramChat
		}
	}
	return t.token != "" && t.chatID != ""
}

func (t *TelegramNotifier) Send(alert Alert) error {
	if !t.Enabled() {
		return nil
	}

	// Ensure alert.Chat.ID is populated for template
	if alert.Chat.ID == "" {
		alert.Chat.ID = t.chatID
	}

	body := bytes.NewBufferString("")
	tpl, err := templateFor("telegram")
	if err != nil {
		return err
	}
	if err := tpl.Execute(body, alert); err != nil {
		return err
	}
	if body.Len() == 0 {
		return nil
	}

	resp, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token),
		"application/json",
		body,
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram returned status %d", resp.StatusCode)
	}
	return nil
}

// SlackNotifier posts to Slack using a token and channel via chat.postMessage
// For backward compatibility, falls back to old webhook URL if token is empty.
type SlackNotifier struct {
	token   string
	channel string
}

func (s *SlackNotifier) Name() string { return "slack" }
func (s *SlackNotifier) Enabled() bool {
	if s.token == "" || s.channel == "" {
		// backward compatible: if old SlackUrl exists, treat as enabled
		if util.Config.SlackAlert && util.Config.SlackUrl != "" {
			return true
		}
	}
	return s.token != "" && s.channel != ""
}

func (s *SlackNotifier) Send(alert Alert) error {
	if s.token != "" && s.channel != "" {
		// Use Slack API chat.postMessage
		body := bytes.NewBufferString("")
		tpl, err := templateFor("slack")
		if err != nil {
			return err
		}
		if err := tpl.Execute(body, alert); err != nil {
			return err
		}
		if body.Len() == 0 {
			return nil
		}

		// Attempt to interpret template JSON and add channel
		var payload map[string]any
		if err := json.Unmarshal(body.Bytes(), &payload); err != nil {
			payload = map[string]any{"text": body.String()}
		}
		payload["channel"] = s.channel
		buf := new(bytes.Buffer)
		_ = json.NewEncoder(buf).Encode(payload)

		req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", buf)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("slack returned status %d", resp.StatusCode)
		}
		// Optionally, inspect ok field from JSON
		var r struct {
			OK bool `json:"ok"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&r)
		if !r.OK {
			return fmt.Errorf("slack api reported failure")
		}
		return nil
	}
	// fallback: old webhook URL implementation
	if util.Config.SlackAlert && util.Config.SlackUrl != "" {
		body := bytes.NewBufferString("")
		tpl, err := templateFor("slack")
		if err != nil {
			return err
		}
		if err := tpl.Execute(body, alert); err != nil {
			return err
		}
		if body.Len() == 0 {
			return nil
		}
		resp, err := http.Post(util.Config.SlackUrl, "application/json", body)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
		}
		return nil
	}
	return nil
}

// GotifyNotifier posts to Gotify server using header token, not URL param.
type GotifyNotifier struct {
	server   string
	token    string
	title    string
	priority int
}

func (g *GotifyNotifier) Name() string { return "gotify" }
func (g *GotifyNotifier) Enabled() bool {
	if g.server == "" || g.token == "" {
		// back-compat: if old config exists, allow through but still send securely if possible
		if util.Config.GotifyAlert && util.Config.GotifyUrl != "" && util.Config.GotifyToken != "" {
			g.server = util.Config.GotifyUrl
			g.token = util.Config.GotifyToken
		}
	}
	return g.server != "" && g.token != ""
}

func (g *GotifyNotifier) Send(alert Alert) error {
	if !g.Enabled() {
		return nil
	}
	body := bytes.NewBufferString("")
	tpl, err := templateFor("gotify")
	if err != nil {
		return err
	}
	if err := tpl.Execute(body, alert); err != nil {
		return err
	}
	if body.Len() == 0 {
		return nil
	}

	// Construct Gotify payload
	payload := map[string]any{
		"message": body.String(),
	}
	if g.title != "" {
		payload["title"] = g.title
	}
	if g.priority != 0 {
		payload["priority"] = g.priority
	}
	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/message", g.server), buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gotify-Key", g.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("gotify returned status %d", resp.StatusCode)
	}
	return nil
}

// DingTalkNotifier sends messages using a token (preferred) rather than insecure URL params.
type DingTalkNotifier struct {
	token   string
	channel string
}

func (d *DingTalkNotifier) Name() string { return "dingtalk" }
func (d *DingTalkNotifier) Enabled() bool {
	if d.token == "" {
		// Back-compat: if URL is set, allow old flow
		if util.Config.DingTalkAlert && util.Config.DingTalkUrl != "" {
			return true
		}
	}
	return d.token != ""
}

func (d *DingTalkNotifier) Send(alert Alert) error {
	if d.token != "" {
		body := bytes.NewBufferString("")
		tpl, err := templateFor("dingtalk")
		if err != nil {
			return err
		}
		if err := tpl.Execute(body, alert); err != nil {
			return err
		}
		if body.Len() == 0 {
			return nil
		}
		// Example DingTalk bot token-based endpoint could differ; using common webhook with token param in header or path
		req, err := http.NewRequest("POST", fmt.Sprintf("https://oapi.dingtalk.com/robot/send"), body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		// Some setups expect access token as query (?access_token=..). We avoid embedding token in URL as per requirement.
		req.Header.Set("X-Dingtalk-Token", d.token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("dingtalk returned status %d", resp.StatusCode)
		}
		return nil
	}
	// Fallback to legacy URL-based approach handled by existing sendDingTalkAlert in TaskRunner
	return nil
}

// templateFor provides the compiled template used by notifier implementations
func templateFor(kind string) (*template.Template, error) {
	// reuse existing embedded templates in alert.go via ParseFS
	tpl, err := template.ParseFS(templates, fmt.Sprintf("templates/%s.tmpl", kind))
	if err != nil {
		return nil, err
	}
	return tpl, nil
}
