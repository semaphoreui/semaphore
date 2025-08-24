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

//go:embed templates/dingtalk.tmpl
var dingtalkTemplate embed.FS

type DingTalkService struct {
	config *util.NotificationServiceConfig
	tmpl   *template.Template
}

type dingtalkAlert struct {
	Name   string
	Author string
	Color  string
	Task   dingtalkTask
}

type dingtalkTask struct {
	ID      string
	URL     string
	Result  string
	Desc    string
	Version string
}

func NewDingTalkService(config *util.NotificationServiceConfig) *DingTalkService {
	tmpl, err := template.ParseFS(dingtalkTemplate, "templates/dingtalk.tmpl")
	if err != nil {
		// Fallback to legacy template system if new template not found
		tmpl = nil
	}
	
	return &DingTalkService{
		config: config,
		tmpl:   tmpl,
	}
}

func (d *DingTalkService) GetName() string {
	return "dingtalk"
}

func (d *DingTalkService) IsConfigured() bool {
	if d.config != nil && d.config.Enabled && d.config.URL != "" {
		return true
	}
	
	// Fallback to legacy configuration
	return util.Config.DingTalkAlert && util.Config.DingTalkUrl != ""
}

func (d *DingTalkService) SupportsProjectOverride() bool {
	return true // DingTalk supports project-specific webhook URLs
}

func (d *DingTalkService) Send(ctx context.Context, notification *Notification) error {
	var webhookURL string
	
	// Use new configuration if available
	if d.config != nil && d.config.Enabled && d.config.URL != "" {
		webhookURL = d.config.URL
		
		// Check for project-specific webhook URL override
		if projectURL, exists := notification.ProjectConfig["dingtalk_url"]; exists {
			if urlStr, ok := projectURL.(string); ok && urlStr != "" {
				webhookURL = urlStr
			}
		}
	} else {
		// Fallback to legacy configuration
		if !util.Config.DingTalkAlert {
			return fmt.Errorf("dingtalk notifications not enabled")
		}
		
		webhookURL = util.Config.DingTalkUrl
	}
	
	if webhookURL == "" {
		return fmt.Errorf("dingtalk webhook URL not configured")
	}
	
	// Create alert data
	alert := dingtalkAlert{
		Name:   notification.TemplateName,
		Author: notification.AuthorName,
		Color:  d.getColor(notification.TaskStatus),
		Task: dingtalkTask{
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
	
	if d.tmpl != nil {
		err = d.tmpl.Execute(body, alert)
	} else {
		// Fallback to legacy template
		err = d.executeLegacyTemplate(body, alert)
	}
	
	if err != nil {
		return fmt.Errorf("failed to generate dingtalk message: %w", err)
	}
	
	if body.Len() == 0 {
		return fmt.Errorf("dingtalk message body is empty")
	}
	
	// Send to DingTalk webhook
	resp, err := http.Post(webhookURL, "application/json", body)
	if err != nil {
		return fmt.Errorf("failed to send dingtalk message: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("dingtalk webhook returned status code: %d", resp.StatusCode)
	}
	
	return nil
}

func (d *DingTalkService) getColor(status task_logger.TaskStatus) string {
	switch status {
	case task_logger.TaskSuccessStatus:
		return "#00EE00"
	case task_logger.TaskFailStatus:
		return "#EE0000"
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

func (d *DingTalkService) executeLegacyTemplate(body *bytes.Buffer, alert dingtalkAlert) error {
	// Simple markdown message format for DingTalk
	message := fmt.Sprintf(`{
		"msgtype": "markdown",
		"markdown": {
			"title": "Task: %s",
			"text": "## Task: %s\n\n**Author:** %s\n\n**Status:** %s\n\n**Message:** %s\n\n[View Task](%s)"
		}
	}`, alert.Name, alert.Name, alert.Author, alert.Task.Result, alert.Task.Desc, alert.Task.URL)
	
	body.WriteString(message)
	return nil
}