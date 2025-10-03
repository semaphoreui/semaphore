-- Add project alert configuration table
CREATE TABLE project_alert_config (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    project_id INTEGER NOT NULL,
    alert_types JSON NOT NULL,
    integrations JSON NOT NULL,
    slack_config JSON,
    teams_config JSON,
    email_config JSON,
    webhook_config JSON,
    discord_config JSON,
    pagerduty_config JSON,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE,
    UNIQUE KEY unique_project_alert_config (project_id)
);

-- Create index for better performance
CREATE INDEX idx_project_alert_config_project_id ON project_alert_config(project_id);
