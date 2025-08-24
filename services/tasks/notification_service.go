package tasks

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// Notifier defines the interface for notification methods
type Notifier interface {
	Send(alert Alert, config util.NotificationConfig) error
	GetTemplateName() string
}

// NotificationService manages all notification methods
type NotificationService struct {
	notifiers map[string]Notifier
	configs   util.NotificationConfigs
}

// NewNotificationService creates a new notification service
func NewNotificationService(configs util.NotificationConfigs) *NotificationService {
	service := &NotificationService{
		notifiers: make(map[string]Notifier),
		configs:   configs,
	}

	// Register all notifiers
	service.notifiers["telegram"] = &TelegramNotifier{}
	service.notifiers["slack"] = &SlackNotifier{}
	service.notifiers["gotify"] = &GotifyNotifier{}
	service.notifiers["dingtalk"] = &DingTalkNotifier{}
	service.notifiers["rocketchat"] = &RocketChatNotifier{}
	service.notifiers["microsoft_teams"] = &MicrosoftTeamsNotifier{}

	return service
}

// SendNotifications sends notifications through all enabled methods
func (ns *NotificationService) SendNotifications(alert Alert, suppressSuccess bool) {
	if suppressSuccess && alert.Task.Result == task_logger.TaskSuccessStatus.Format() {
		return
	}

	// Send through all enabled notification methods
	if ns.configs.Telegram != nil && ns.configs.Telegram.Enabled {
		if err := ns.notifiers["telegram"].Send(alert, *ns.configs.Telegram); err != nil {
			util.LogError(fmt.Errorf("telegram notification failed: %w", err))
		}
	}

	if ns.configs.Slack != nil && ns.configs.Slack.Enabled {
		if err := ns.notifiers["slack"].Send(alert, *ns.configs.Slack); err != nil {
			util.LogError(fmt.Errorf("slack notification failed: %w", err))
		}
	}

	if ns.configs.Gotify != nil && ns.configs.Gotify.Enabled {
		if err := ns.notifiers["gotify"].Send(alert, *ns.configs.Gotify); err != nil {
			util.LogError(fmt.Errorf("gotify notification failed: %w", err))
		}
	}

	if ns.configs.DingTalk != nil && ns.configs.DingTalk.Enabled {
		if err := ns.notifiers["dingtalk"].Send(alert, *ns.configs.DingTalk); err != nil {
			util.LogError(fmt.Errorf("dingtalk notification failed: %w", err))
		}
	}

	if ns.configs.RocketChat != nil && ns.configs.RocketChat.Enabled {
		if err := ns.notifiers["rocketchat"].Send(alert, *ns.configs.RocketChat); err != nil {
			util.LogError(fmt.Errorf("rocketchat notification failed: %w", err))
		}
	}

	if ns.configs.MicrosoftTeams != nil && ns.configs.MicrosoftTeams.Enabled {
		if err := ns.notifiers["microsoft_teams"].Send(alert, *ns.configs.MicrosoftTeams); err != nil {
			util.LogError(fmt.Errorf("microsoft teams notification failed: %w", err))
		}
	}
}

// renderTemplate renders the notification template for the given notifier
func renderTemplate(notifier Notifier, alert Alert) (string, error) {
	tpl, err := template.ParseFS(templates, fmt.Sprintf("templates/%s.tmpl", notifier.GetTemplateName()))
	if err != nil {
		return "", fmt.Errorf("can't parse %s template: %w", notifier.GetTemplateName(), err)
	}

	body := bytes.NewBufferString("")
	if err := tpl.Execute(body, alert); err != nil {
		return "", fmt.Errorf("can't generate %s template: %w", notifier.GetTemplateName(), err)
	}

	if body.Len() == 0 {
		return "", fmt.Errorf("buffer for %s alert is empty", notifier.GetTemplateName())
	}

	return body.String(), nil
}

// TelegramNotifier implements Telegram notifications
type TelegramNotifier struct{}

func (tn *TelegramNotifier) Send(alert Alert, config util.NotificationConfig) error {
	// Set chat ID from channel if provided, otherwise use the one in alert
	if config.Channel != "" {
		alert.Chat.ID = config.Channel
	}

	body, err := renderTemplate(tn, alert)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.Token),
		"application/json",
		bytes.NewBufferString(body),
	)

	if err != nil {
		return fmt.Errorf("failed to send telegram notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (tn *TelegramNotifier) GetTemplateName() string {
	return "telegram"
}

// SlackNotifier implements Slack notifications
type SlackNotifier struct{}

func (sn *SlackNotifier) Send(alert Alert, config util.NotificationConfig) error {
	body, err := renderTemplate(sn, alert)
	if err != nil {
		return err
	}

	url := config.URL
	if url == "" {
		return fmt.Errorf("slack URL is required")
	}

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("slack API returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (sn *SlackNotifier) GetTemplateName() string {
	return "slack"
}

// GotifyNotifier implements Gotify notifications
type GotifyNotifier struct{}

func (gn *GotifyNotifier) Send(alert Alert, config util.NotificationConfig) error {
	body, err := renderTemplate(gn, alert)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/message?token=%s", config.URL, config.Token)
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("failed to send gotify notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("gotify API returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (gn *GotifyNotifier) GetTemplateName() string {
	return "gotify"
}

// DingTalkNotifier implements DingTalk notifications
type DingTalkNotifier struct{}

func (dn *DingTalkNotifier) Send(alert Alert, config util.NotificationConfig) error {
	body, err := renderTemplate(dn, alert)
	if err != nil {
		return err
	}

	url := config.URL
	if url == "" {
		return fmt.Errorf("dingtalk URL is required")
	}

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("failed to send dingtalk notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("dingtalk API returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (dn *DingTalkNotifier) GetTemplateName() string {
	return "dingtalk"
}

// RocketChatNotifier implements RocketChat notifications
type RocketChatNotifier struct{}

func (rn *RocketChatNotifier) Send(alert Alert, config util.NotificationConfig) error {
	body, err := renderTemplate(rn, alert)
	if err != nil {
		return err
	}

	url := config.URL
	if url == "" {
		return fmt.Errorf("rocketchat URL is required")
	}

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("failed to send rocketchat notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("rocketchat API returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (rn *RocketChatNotifier) GetTemplateName() string {
	return "rocketchat"
}

// MicrosoftTeamsNotifier implements Microsoft Teams notifications
type MicrosoftTeamsNotifier struct{}

func (mn *MicrosoftTeamsNotifier) Send(alert Alert, config util.NotificationConfig) error {
	body, err := renderTemplate(mn, alert)
	if err != nil {
		return err
	}

	url := config.URL
	if url == "" {
		return fmt.Errorf("microsoft teams URL is required")
	}

	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("failed to send microsoft teams notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return fmt.Errorf("microsoft teams API returned status code: %d", resp.StatusCode)
	}

	return nil
}

func (mn *MicrosoftTeamsNotifier) GetTemplateName() string {
	return "microsoft-teams"
}