package notifications

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

//go:embed templates/slack.tmpl
var slackTemplate embed.FS

type SlackService struct {
	config *util.NotificationServiceConfig
	tmpl   *template.Template
}

type slackAlert struct {
	Name   string
	Author string
	Color  string
	Task   slackTask
}

type slackTask struct {
	ID      string
	URL     string
	Result  string
	Desc    string
	Version string
}

func NewSlackService(config *util.NotificationServiceConfig) *SlackService {
	tmpl, err := template.ParseFS(slackTemplate, "templates/slack.tmpl")
	if err != nil {
		// Fallback to legacy template system if new template not found
		tmpl = nil
	}
	
	return &SlackService{
		config: config,
		tmpl:   tmpl,
	}
}

func (s *SlackService) GetName() string {
	return "slack"
}

func (s *SlackService) IsConfigured() bool {
	if s.config != nil && s.config.Enabled && s.config.URL != "" {
		return true
	}
	
	// Fallback to legacy configuration
	return util.Config.SlackAlert && util.Config.SlackUrl != ""
}

func (s *SlackService) SupportsProjectOverride() bool {
	return true // Slack supports project-specific webhook URLs
}

func (s *SlackService) Send(ctx context.Context, notification *Notification) error {
	var webhookURL string
	
	// Use new configuration if available
	if s.config != nil && s.config.Enabled && s.config.URL != "" {
		webhookURL = s.config.URL
		
		// Check for project-specific webhook URL override
		if projectURL, exists := notification.ProjectConfig["slack_url"]; exists {
			if urlStr, ok := projectURL.(string); ok && urlStr != "" {
				webhookURL = urlStr
			}
		}
	} else {
		// Fallback to legacy configuration
		if !util.Config.SlackAlert {
			return fmt.Errorf("slack notifications not enabled")
		}
		
		webhookURL = util.Config.SlackUrl
	}
	
	if webhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}
	
	// Create alert data
	alert := slackAlert{
		Name:   notification.TemplateName,
		Author: notification.AuthorName,
		Color:  s.getColor(notification.TaskStatus),
		Task: slackTask{
			ID:      strconv.Itoa(notification.TaskID),
			URL:     notification.TaskURL,
			Result:  notification.TaskStatus.Format(),
			Version: notification.TaskVersion,
			Desc:    notification.TaskMessage,
		},
	}
	
	// Generate message body
	body := bytes.NewBufferString("")
	var err error
	
	if s.tmpl != nil {
		err = s.tmpl.Execute(body, alert)
	} else {
		// Fallback to legacy template
		err = s.executeLegacyTemplate(body, alert)
	}
	
	if err != nil {
		return fmt.Errorf("failed to generate slack message: %w", err)
	}
	
	if body.Len() == 0 {
		return fmt.Errorf("slack message body is empty")
	}
	
	// Send to Slack webhook
	resp, err := http.Post(webhookURL, "application/json", body)
	if err != nil {
		return fmt.Errorf("failed to send slack message: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("slack webhook returned status code: %d", resp.StatusCode)
	}
	
	return nil
}

func (s *SlackService) getColor(status task_logger.TaskStatus) string {
	switch status {
	case task_logger.TaskSuccessStatus:
		return "good"
	case task_logger.TaskFailStatus:
		return "danger"
	case task_logger.TaskRunningStatus:
		return "#333CFF"
	case task_logger.TaskWaitingStatus:
		return "#FFFC33"
	case task_logger.TaskStoppingStatus:
		return "#BEBEBE"
	case task_logger.TaskStoppedStatus:
		return "#5B5B5B"
	default:
		return ""
	}
}

func (s *SlackService) executeLegacyTemplate(body *bytes.Buffer, alert slackAlert) error {
	// Simple fallback message format for Slack
	message := fmt.Sprintf(`{
		"attachments": [
			{
				"color": "%s",
				"title": "Task: %s",
				"fields": [
					{
						"title": "Author",
						"value": "%s",
						"short": true
					},
					{
						"title": "Status",
						"value": "%s",
						"short": true
					},
					{
						"title": "Message",
						"value": "%s",
						"short": false
					}
				],
				"actions": [
					{
						"type": "button",
						"text": "View Task",
						"url": "%s"
					}
				]
			}
		]
	}`, alert.Color, alert.Name, alert.Author, alert.Task.Result, alert.Task.Desc, alert.Task.URL)
	
	body.WriteString(message)
	return nil
}