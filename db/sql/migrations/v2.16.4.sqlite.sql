-- Create task_files table for storing task artifacts (SQLite version)
CREATE TABLE task_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    project_id INTEGER NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type VARCHAR(100),
    checksum VARCHAR(64),
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    description TEXT,
    FOREIGN KEY (task_id) REFERENCES task(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_task_files_task_id ON task_files(task_id);
CREATE INDEX idx_task_files_project_id ON task_files(project_id);
CREATE INDEX idx_task_files_created ON task_files(created);
