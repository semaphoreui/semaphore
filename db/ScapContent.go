package db

import (
	"encoding/json"
	"time"
)

// ScapContent represents a SCAP DataStream file
type ScapContent struct {
	ID          int       `db:"id" json:"id"`
	ProjectID   int       `db:"project_id" json:"project_id"`
	Name        string    `db:"name" json:"name"`
	Source      *string   `db:"source" json:"source"`
	DsXmlPath   *string   `db:"ds_xml_path" json:"ds_xml_path"`
	UploadedBy  int       `db:"uploaded_by" json:"uploaded_by"`
	Created     time.Time `db:"created" json:"created"`
}

// ScapProfile represents a profile discovered from a SCAP DataStream
type ScapProfile struct {
	ID          int     `db:"id" json:"id"`
	ContentID   int     `db:"content_id" json:"content_id"`
	ProfileID   string  `db:"profile_id" json:"profile_id"`
	Title       string  `db:"title" json:"title"`
	Severity    *string `db:"severity" json:"severity"`
	Description *string `db:"description" json:"description"`
}

// CompliancePolicy represents a compliance scanning policy
type CompliancePolicy struct {
	ID        int       `db:"id" json:"id"`
	ProjectID int       `db:"project_id" json:"project_id"`
	Name      string    `db:"name" json:"name"`
	ContentID int       `db:"content_id" json:"content_id"`
	ProfileID string    `db:"profile_id" json:"profile_id"`
	ScheduleID *int     `db:"schedule_id" json:"schedule_id"`
	Attrs     *string   `db:"attrs" json:"attrs"`
	Created   time.Time `db:"created" json:"created"`
	CreatedBy int       `db:"created_by" json:"created_by"`
}

// GetAttrs returns the parsed JSON attributes
func (cp *CompliancePolicy) GetAttrs() (map[string]interface{}, error) {
	if cp.Attrs == nil {
		return make(map[string]interface{}), nil
	}
	
	var attrs map[string]interface{}
	err := json.Unmarshal([]byte(*cp.Attrs), &attrs)
	return attrs, err
}

// SetAttrs sets the JSON attributes from a map
func (cp *CompliancePolicy) SetAttrs(attrs map[string]interface{}) error {
	if attrs == nil {
		cp.Attrs = nil
		return nil
	}
	
	data, err := json.Marshal(attrs)
	if err != nil {
		return err
	}
	
	attrsStr := string(data)
	cp.Attrs = &attrsStr
	return nil
}

// PolicyAssignment represents assignment of a policy to targets
type PolicyAssignment struct {
	ID         int       `db:"id" json:"id"`
	PolicyID   int       `db:"policy_id" json:"policy_id"`
	TargetType string    `db:"target_type" json:"target_type"` // 'inventory', 'host', 'group'
	TargetID   int       `db:"target_id" json:"target_id"`
	Created    time.Time `db:"created" json:"created"`
}

// ComplianceScan represents a compliance scan execution
type ComplianceScan struct {
	ID          int        `db:"id" json:"id"`
	ProjectID   int        `db:"project_id" json:"project_id"`
	PolicyID    int        `db:"policy_id" json:"policy_id"`
	InitiatedBy int        `db:"initiated_by" json:"initiated_by"`
	Started     time.Time  `db:"started" json:"started"`
	Ended       *time.Time `db:"ended" json:"ended"`
	Status      string     `db:"status" json:"status"` // 'running', 'completed', 'failed', 'cancelled'
}

// ComplianceReport represents scan results for a single host
type ComplianceReport struct {
	ID       int       `db:"id" json:"id"`
	ScanID   int       `db:"scan_id" json:"scan_id"`
	Host     string    `db:"host" json:"host"`
	Result   string    `db:"result" json:"result"` // 'pass', 'fail', 'error', 'notapplicable'
	ArfPath  *string   `db:"arf_path" json:"arf_path"`
	Summary  *string   `db:"summary" json:"summary"`
	Created  time.Time `db:"created" json:"created"`
}

// GetSummary returns the parsed JSON summary
func (cr *ComplianceReport) GetSummary() (map[string]interface{}, error) {
	if cr.Summary == nil {
		return make(map[string]interface{}), nil
	}
	
	var summary map[string]interface{}
	err := json.Unmarshal([]byte(*cr.Summary), &summary)
	return summary, err
}

// SetSummary sets the JSON summary from a map
func (cr *ComplianceReport) SetSummary(summary map[string]interface{}) error {
	if summary == nil {
		cr.Summary = nil
		return nil
	}
	
	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	
	summaryStr := string(data)
	cr.Summary = &summaryStr
	return nil
}

// ComplianceRuleResult represents individual rule results
type ComplianceRuleResult struct {
	ID          int     `db:"id" json:"id"`
	ReportID    int     `db:"report_id" json:"report_id"`
	RuleID      string  `db:"rule_id" json:"rule_id"`
	Severity    *string `db:"severity" json:"severity"`
	Result      string  `db:"result" json:"result"` // 'pass', 'fail', 'error', 'notapplicable', 'notchecked', 'notselected'
	Ident       *string `db:"ident" json:"ident"`
	Description *string `db:"description" json:"description"`
}
