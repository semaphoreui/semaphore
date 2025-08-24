package notifications

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

//go:embed templates/telegram.tmpl
var telegramTemplate embed.FS

type TelegramService struct {
	config *util.NotificationServiceConfig
	tmpl   *template.Template
}

type telegramAlert struct {
	Name   string
	Author string
	Color  string
	Task   telegramTask
	Chat   telegramChat
}

type telegramTask struct {
	ID      string
	URL     string
	Result  string
	Desc    string
	Version string
}

type telegramChat struct {
	ID string
}

func NewTelegramService(config *util.NotificationServiceConfig) *TelegramService {
	tmpl, err := template.ParseFS(telegramTemplate, "templates/telegram.tmpl")
	if err != nil {
		// Fallback to legacy template system if new template not found
		tmpl = nil
	}
	
	return &TelegramService{
		config: config,
		tmpl:   tmpl,
	}
}

func (t *TelegramService) GetName() string {
	return "telegram"
}

func (t *TelegramService) IsConfigured() bool {
	if t.config != nil && t.config.Enabled && t.config.Token != "" {
		return true
	}
	
	// Fallback to legacy configuration
	return util.Config.TelegramAlert && util.Config.TelegramToken != ""
}

func (t *TelegramService) SupportsProjectOverride() bool {
	return true // Telegram supports project-specific chat IDs
}

func (t *TelegramService) Send(ctx context.Context, notification *Notification) error {
	var token, chatID string
	
	// Use new configuration if available
	if t.config != nil && t.config.Enabled && t.config.Token != "" {
		token = t.config.Token
		chatID = t.config.Channel
		
		// Check for project-specific chat ID override
		if projectChatID, exists := notification.ProjectConfig["telegram_chat"]; exists {
			if chatIDStr, ok := projectChatID.(string); ok && chatIDStr != "" {
				chatID = chatIDStr
			}
		}
	} else {
		// Fallback to legacy configuration
		if !util.Config.TelegramAlert {
			return fmt.Errorf("telegram notifications not enabled")
		}
		
		token = util.Config.TelegramToken
		chatID = util.Config.TelegramChat
		
		// Legacy project chat override (from project.AlertChat)
		if len(notification.Recipients) > 0 {
			chatID = notification.Recipients[0]
		}
	}
	
	if token == "" {
		return fmt.Errorf("telegram token not configured")
	}
	
	if chatID == "" {
		return fmt.Errorf("telegram chat ID not configured")
	}
	
	// Create alert data
	alert := telegramAlert{
		Name:   notification.TemplateName,
		Author: notification.AuthorName,
		Color:  t.getColor(notification.TaskStatus),
		Task: telegramTask{
			ID:      strconv.Itoa(notification.TaskID),
			URL:     notification.TaskURL,
			Result:  notification.TaskStatus.Format(),
			Version: notification.TaskVersion,
			Desc:    notification.TaskMessage,
		},
		Chat: telegramChat{
			ID: chatID,
		},
	}
	
	// Generate message body
	body := bytes.NewBufferString("")
	var err error
	
	if t.tmpl != nil {
		err = t.tmpl.Execute(body, alert)
	} else {
		// Fallback to legacy template
		err = t.executeLegacyTemplate(body, alert)
	}
	
	if err != nil {
		return fmt.Errorf("failed to generate telegram message: %w", err)
	}
	
	if body.Len() == 0 {
		return fmt.Errorf("telegram message body is empty")
	}
	
	// Send to Telegram API
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	
	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram API returned status code: %d", resp.StatusCode)
	}
	
	return nil
}

func (t *TelegramService) getColor(status task_logger.TaskStatus) string {
	// Telegram doesn't use colors in the same way as Slack
	return ""
}

func (t *TelegramService) executeLegacyTemplate(body *bytes.Buffer, alert telegramAlert) error {
	// Simple JSON message format for Telegram
	message := map[string]interface{}{
		"chat_id":    alert.Chat.ID,
		"text":       fmt.Sprintf("📋 *%s*\n👤 Author: %s\n📊 Status: %s\n📝 Message: %s\n🔗 [View Task](%s)", 
			alert.Name, alert.Author, alert.Task.Result, alert.Task.Desc, alert.Task.URL),
		"parse_mode": "Markdown",
	}
	
	return json.NewEncoder(body).Encode(message)
}