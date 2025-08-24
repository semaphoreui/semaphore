package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

func TestNewNotificationService(t *testing.T) {
	configs := util.NotificationConfigs{
		Telegram: &util.NotificationConfig{
			Enabled: true,
			Token:   "test_token",
			Channel: "test_channel",
		},
		Slack: &util.NotificationConfig{
			Enabled: true,
			URL:     "https://hooks.slack.com/test",
		},
	}

	service := NewNotificationService(configs)

	if service == nil {
		t.Fatal("NewNotificationService returned nil")
	}

	if len(service.notifiers) != 6 {
		t.Errorf("Expected 6 notifiers, got %d", len(service.notifiers))
	}

	// Check if all expected notifiers are registered
	expectedNotifiers := []string{"telegram", "slack", "gotify", "dingtalk", "rocketchat", "microsoft_teams"}
	for _, name := range expectedNotifiers {
		if _, exists := service.notifiers[name]; !exists {
			t.Errorf("Expected notifier %s not found", name)
		}
	}
}

func TestTelegramNotifier(t *testing.T) {
	notifier := &TelegramNotifier{}
	
	if notifier.GetTemplateName() != "telegram" {
		t.Errorf("Expected template name 'telegram', got '%s'", notifier.GetTemplateName())
	}
}

func TestSlackNotifier(t *testing.T) {
	notifier := &SlackNotifier{}
	
	if notifier.GetTemplateName() != "slack" {
		t.Errorf("Expected template name 'slack', got '%s'", notifier.GetTemplateName())
	}
}

func TestGotifyNotifier(t *testing.T) {
	notifier := &GotifyNotifier{}
	
	if notifier.GetTemplateName() != "gotify" {
		t.Errorf("Expected template name 'gotify', got '%s'", notifier.GetTemplateName())
	}
}

func TestDingTalkNotifier(t *testing.T) {
	notifier := &DingTalkNotifier{}
	
	if notifier.GetTemplateName() != "dingtalk" {
		t.Errorf("Expected template name 'dingtalk', got '%s'", notifier.GetTemplateName())
	}
}

func TestRocketChatNotifier(t *testing.T) {
	notifier := &RocketChatNotifier{}
	
	if notifier.GetTemplateName() != "rocketchat" {
		t.Errorf("Expected template name 'rocketchat', got '%s'", notifier.GetTemplateName())
	}
}

func TestMicrosoftTeamsNotifier(t *testing.T) {
	notifier := &MicrosoftTeamsNotifier{}
	
	if notifier.GetTemplateName() != "microsoft-teams" {
		t.Errorf("Expected template name 'microsoft-teams', got '%s'", notifier.GetTemplateName())
	}
}

func TestSendNotificationsWithDisabledNotifiers(t *testing.T) {
	configs := util.NotificationConfigs{
		Telegram: &util.NotificationConfig{
			Enabled: false,
			Token:   "test_token",
			Channel: "test_channel",
		},
		Slack: &util.NotificationConfig{
			Enabled: false,
			URL:     "https://hooks.slack.com/test",
		},
	}

	service := NewNotificationService(configs)

	alert := Alert{
		Name:   "Test Template",
		Author: "Test User",
		Color:  "",
		Task: alertTask{
			ID:      "123",
			URL:     "http://test.com/task/123",
			Result:  task_logger.TaskSuccessStatus.Format(),
			Version: "v1.0.0",
			Desc:    "Test task description",
		},
		Chat: alertChat{
			ID: "test_chat",
		},
	}

	// This should not panic and should not send any notifications
	// since all notifiers are disabled
	service.SendNotifications(alert, false)
}