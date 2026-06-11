CREATE TABLE project__notification (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE
);

CREATE TABLE project__template_notification (
    project_id INTEGER NOT NULL,
    template_id INTEGER NOT NULL,
    notification_id INTEGER NOT NULL,
    PRIMARY KEY (project_id, template_id, notification_id)
);
