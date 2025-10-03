-- OpenSCAP Compliance Tables Migration v2.17.0

-- SCAP Contents table for storing SCAP DataStream files
CREATE TABLE scap_contents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    source VARCHAR(500),
    ds_xml_path VARCHAR(500),
    uploaded_by INTEGER NOT NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (project_id) REFERENCES project (id) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by) REFERENCES user (id) ON DELETE CASCADE,
    
    UNIQUE (project_id, name)
);

-- SCAP Profiles table for discovered profiles from DataStreams
CREATE TABLE scap_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content_id INTEGER NOT NULL,
    profile_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    severity VARCHAR(50),
    description TEXT,
    
    FOREIGN KEY (content_id) REFERENCES scap_contents (id) ON DELETE CASCADE,
    
    UNIQUE (content_id, profile_id)
);

-- Compliance Policies table
CREATE TABLE compliance_policies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    content_id INTEGER NOT NULL,
    profile_id VARCHAR(255) NOT NULL,
    schedule_id INTEGER NULL,
    attrs TEXT, -- JSON attributes for policy configuration
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER NOT NULL,
    
    FOREIGN KEY (project_id) REFERENCES project (id) ON DELETE CASCADE,
    FOREIGN KEY (content_id) REFERENCES scap_contents (id) ON DELETE CASCADE,
    FOREIGN KEY (schedule_id) REFERENCES schedule (id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES user (id) ON DELETE CASCADE,
    
    UNIQUE (project_id, name)
);

-- Policy Assignments table for targeting specific hosts/inventories
CREATE TABLE policy_assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    policy_id INTEGER NOT NULL,
    target_type VARCHAR(50) NOT NULL, -- 'inventory', 'host', 'group'
    target_id INTEGER NOT NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (policy_id) REFERENCES compliance_policies (id) ON DELETE CASCADE,
    
    UNIQUE (policy_id, target_type, target_id)
);

-- Compliance Scans table for tracking scan executions
CREATE TABLE compliance_scans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    policy_id INTEGER NOT NULL,
    initiated_by INTEGER NOT NULL,
    started DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended DATETIME NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'running', -- 'running', 'completed', 'failed', 'cancelled'
    
    FOREIGN KEY (project_id) REFERENCES project (id) ON DELETE CASCADE,
    FOREIGN KEY (policy_id) REFERENCES compliance_policies (id) ON DELETE CASCADE,
    FOREIGN KEY (initiated_by) REFERENCES user (id) ON DELETE CASCADE
);

-- Compliance Reports table for individual host scan results
CREATE TABLE compliance_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_id INTEGER NOT NULL,
    host VARCHAR(255) NOT NULL,
    result VARCHAR(50) NOT NULL, -- 'pass', 'fail', 'error', 'notapplicable'
    arf_path VARCHAR(500),
    summary TEXT, -- JSON summary of scan results
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (scan_id) REFERENCES compliance_scans (id) ON DELETE CASCADE,
    
    UNIQUE (scan_id, host)
);

-- Compliance Rule Results table for detailed rule-level results
CREATE TABLE compliance_rule_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL,
    rule_id VARCHAR(255) NOT NULL,
    severity VARCHAR(50),
    result VARCHAR(50) NOT NULL, -- 'pass', 'fail', 'error', 'notapplicable', 'notchecked', 'notselected'
    ident VARCHAR(500),
    description TEXT,
    
    FOREIGN KEY (report_id) REFERENCES compliance_reports (id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_scap_contents_project_id ON scap_contents (project_id);
CREATE INDEX idx_scap_profiles_content_id ON scap_profiles (content_id);
CREATE INDEX idx_compliance_policies_project_id ON compliance_policies (project_id);
CREATE INDEX idx_compliance_policies_content_id ON compliance_policies (content_id);
CREATE INDEX idx_policy_assignments_policy_id ON policy_assignments (policy_id);
CREATE INDEX idx_compliance_scans_project_id ON compliance_scans (project_id);
CREATE INDEX idx_compliance_scans_policy_id ON compliance_scans (policy_id);
CREATE INDEX idx_compliance_scans_status ON compliance_scans (status);
CREATE INDEX idx_compliance_reports_scan_id ON compliance_reports (scan_id);
CREATE INDEX idx_compliance_reports_host ON compliance_reports (host);
CREATE INDEX idx_compliance_reports_result ON compliance_reports (result);
CREATE INDEX idx_compliance_rule_results_report_id ON compliance_rule_results (report_id);
