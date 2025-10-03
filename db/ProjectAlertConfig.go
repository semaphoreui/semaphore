package db

import (
	"encoding/json"
	"time"

	"github.com/go-gorp/gorp/v3"
)

// ProjectAlertConfig represents alert configuration for a project
type ProjectAlertConfig struct {
	ID             int       `db:"id" json:"id"`
	ProjectID      int       `db:"project_id" json:"project_id"`
	AlertTypes     string    `db:"alert_types" json:"alert_types"`
	Integrations   string    `db:"integrations" json:"integrations"`
	SlackConfig    *string   `db:"slack_config" json:"slack_config,omitempty"`
	TeamsConfig    *string   `db:"teams_config" json:"teams_config,omitempty"`
	EmailConfig    *string   `db:"email_config" json:"email_config,omitempty"`
	WebhookConfig  *string   `db:"webhook_config" json:"webhook_config,omitempty"`
	DiscordConfig  *string   `db:"discord_config" json:"discord_config,omitempty"`
	PagerDutyConfig *string  `db:"pagerduty_config" json:"pagerduty_config,omitempty"`
	Created        time.Time `db:"created" json:"created"`
	Updated        time.Time `db:"updated" json:"updated"`
}

// AlertTypes represents the alert types configuration
type AlertTypes struct {
	TaskFailure    bool `json:"taskFailure"`
	TaskSuccess    bool `json:"taskSuccess"`
	ProjectCreated bool `json:"projectCreated"`
	UserLogin      bool `json:"userLogin"`
	SystemError    bool `json:"systemError"`
	SecurityEvent  bool `json:"securityEvent"`
}

// Integrations represents the integrations configuration
type Integrations struct {
	Slack     bool `json:"slack"`
	Teams     bool `json:"teams"`
	Email     bool `json:"email"`
	Webhook   bool `json:"webhook"`
	Discord   bool `json:"discord"`
	PagerDuty bool `json:"pagerduty"`
}

// SlackConfig represents Slack configuration
type SlackConfig struct {
	WebhookURL string `json:"webhookUrl"`
	Channel    string `json:"channel"`
}

// TeamsConfig represents Microsoft Teams configuration
type TeamsConfig struct {
	WebhookURL string `json:"webhookUrl"`
}

// EmailConfig represents Email configuration
type EmailConfig struct {
	Recipients string `json:"recipients"`
	Subject    string `json:"subject"`
}

// WebhookConfig represents Webhook configuration
type WebhookConfig struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// DiscordConfig represents Discord configuration
type DiscordConfig struct {
	WebhookURL string `json:"webhookUrl"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatarUrl"`
}

// PagerDutyConfig represents PagerDuty configuration
type PagerDutyConfig struct {
	IntegrationKey string `json:"integrationKey"`
	ServiceID      string `json:"serviceId"`
	Severity       string `json:"severity"`
}

// GetAlertTypes returns the parsed alert types
func (pac *ProjectAlertConfig) GetAlertTypes() (*AlertTypes, error) {
	var alertTypes AlertTypes
	if err := json.Unmarshal([]byte(pac.AlertTypes), &alertTypes); err != nil {
		return nil, err
	}
	return &alertTypes, nil
}

// SetAlertTypes sets the alert types from a struct
func (pac *ProjectAlertConfig) SetAlertTypes(alertTypes *AlertTypes) error {
	data, err := json.Marshal(alertTypes)
	if err != nil {
		return err
	}
	pac.AlertTypes = string(data)
	return nil
}

// GetIntegrations returns the parsed integrations
func (pac *ProjectAlertConfig) GetIntegrations() (*Integrations, error) {
	var integrations Integrations
	if err := json.Unmarshal([]byte(pac.Integrations), &integrations); err != nil {
		return nil, err
	}
	return &integrations, nil
}

// SetIntegrations sets the integrations from a struct
func (pac *ProjectAlertConfig) SetIntegrations(integrations *Integrations) error {
	data, err := json.Marshal(integrations)
	if err != nil {
		return err
	}
	pac.Integrations = string(data)
	return nil
}

// GetSlackConfig returns the parsed Slack configuration
func (pac *ProjectAlertConfig) GetSlackConfig() (*SlackConfig, error) {
	if pac.SlackConfig == nil {
		return nil, nil
	}
	var config SlackConfig
	if err := json.Unmarshal([]byte(*pac.SlackConfig), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetSlackConfig sets the Slack configuration from a struct
func (pac *ProjectAlertConfig) SetSlackConfig(config *SlackConfig) error {
	if config == nil {
		pac.SlackConfig = nil
		return nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configStr := string(data)
	pac.SlackConfig = &configStr
	return nil
}

// GetTeamsConfig returns the parsed Teams configuration
func (pac *ProjectAlertConfig) GetTeamsConfig() (*TeamsConfig, error) {
	if pac.TeamsConfig == nil {
		return nil, nil
	}
	var config TeamsConfig
	if err := json.Unmarshal([]byte(*pac.TeamsConfig), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetTeamsConfig sets the Teams configuration from a struct
func (pac *ProjectAlertConfig) SetTeamsConfig(config *TeamsConfig) error {
	if config == nil {
		pac.TeamsConfig = nil
		return nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configStr := string(data)
	pac.TeamsConfig = &configStr
	return nil
}

// GetEmailConfig returns the parsed Email configuration
func (pac *ProjectAlertConfig) GetEmailConfig() (*EmailConfig, error) {
	if pac.EmailConfig == nil {
		return nil, nil
	}
	var config EmailConfig
	if err := json.Unmarshal([]byte(*pac.EmailConfig), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetEmailConfig sets the Email configuration from a struct
func (pac *ProjectAlertConfig) SetEmailConfig(config *EmailConfig) error {
	if config == nil {
		pac.EmailConfig = nil
		return nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configStr := string(data)
	pac.EmailConfig = &configStr
	return nil
}

// GetWebhookConfig returns the parsed Webhook configuration
func (pac *ProjectAlertConfig) GetWebhookConfig() (*WebhookConfig, error) {
	if pac.WebhookConfig == nil {
		return nil, nil
	}
	var config WebhookConfig
	if err := json.Unmarshal([]byte(*pac.WebhookConfig), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetWebhookConfig sets the Webhook configuration from a struct
func (pac *ProjectAlertConfig) SetWebhookConfig(config *WebhookConfig) error {
	if config == nil {
		pac.WebhookConfig = nil
		return nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configStr := string(data)
	pac.WebhookConfig = &configStr
	return nil
}

// GetDiscordConfig returns the parsed Discord configuration
func (pac *ProjectAlertConfig) GetDiscordConfig() (*DiscordConfig, error) {
	if pac.DiscordConfig == nil {
		return nil, nil
	}
	var config DiscordConfig
	if err := json.Unmarshal([]byte(*pac.DiscordConfig), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetDiscordConfig sets the Discord configuration from a struct
func (pac *ProjectAlertConfig) SetDiscordConfig(config *DiscordConfig) error {
	if config == nil {
		pac.DiscordConfig = nil
		return nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configStr := string(data)
	pac.DiscordConfig = &configStr
	return nil
}

// GetPagerDutyConfig returns the parsed PagerDuty configuration
func (pac *ProjectAlertConfig) GetPagerDutyConfig() (*PagerDutyConfig, error) {
	if pac.PagerDutyConfig == nil {
		return nil, nil
	}
	var config PagerDutyConfig
	if err := json.Unmarshal([]byte(*pac.PagerDutyConfig), &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetPagerDutyConfig sets the PagerDuty configuration from a struct
func (pac *ProjectAlertConfig) SetPagerDutyConfig(config *PagerDutyConfig) error {
	if config == nil {
		pac.PagerDutyConfig = nil
		return nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	configStr := string(data)
	pac.PagerDutyConfig = &configStr
	return nil
}

// PreInsert sets the created timestamp before inserting
func (pac *ProjectAlertConfig) PreInsert(s gorp.SqlExecutor) error {
	pac.Created = time.Now()
	pac.Updated = time.Now()
	return nil
}

// PreUpdate sets the updated timestamp before updating
func (pac *ProjectAlertConfig) PreUpdate(s gorp.SqlExecutor) error {
	pac.Updated = time.Now()
	return nil
}
