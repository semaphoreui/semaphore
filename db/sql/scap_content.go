package sql

import (
	"fmt"
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Masterminds/squirrel"
)

// ScapContent methods

func (d *SqlDb) CreateScapContent(content *db.ScapContent) error {
	return d.Sql().Insert(content)
}

func (d *SqlDb) GetScapContent(id int) (*db.ScapContent, error) {
	var content db.ScapContent
	err := d.Sql().SelectOne(&content, "SELECT * FROM scap_contents WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &content, nil
}

func (d *SqlDb) GetScapContentsByProject(projectID int) ([]*db.ScapContent, error) {
	var contents []*db.ScapContent
	_, err := d.Sql().Select(&contents, "SELECT * FROM scap_contents WHERE project_id = ? ORDER BY created DESC", projectID)
	return contents, err
}

func (d *SqlDb) UpdateScapContent(content *db.ScapContent) error {
	_, err := d.Sql().Update(content)
	return err
}

func (d *SqlDb) DeleteScapContent(id int) error {
	_, err := d.Sql().Exec("DELETE FROM scap_contents WHERE id = ?", id)
	return err
}

// ScapProfile methods

func (d *SqlDb) CreateScapProfile(profile *db.ScapProfile) error {
	return d.Sql().Insert(profile)
}

func (d *SqlDb) GetScapProfilesByContent(contentID int) ([]*db.ScapProfile, error) {
	var profiles []*db.ScapProfile
	_, err := d.Sql().Select(&profiles, "SELECT * FROM scap_profiles WHERE content_id = ? ORDER BY title", contentID)
	return profiles, err
}

func (d *SqlDb) GetScapProfile(contentID int, profileID string) (*db.ScapProfile, error) {
	var profile db.ScapProfile
	err := d.Sql().SelectOne(&profile, "SELECT * FROM scap_profiles WHERE content_id = ? AND profile_id = ?", contentID, profileID)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (d *SqlDb) DeleteScapProfilesByContent(contentID int) error {
	_, err := d.Sql().Exec("DELETE FROM scap_profiles WHERE content_id = ?", contentID)
	return err
}

// CompliancePolicy methods

func (d *SqlDb) CreateCompliancePolicy(policy *db.CompliancePolicy) error {
	return d.Sql().Insert(policy)
}

func (d *SqlDb) GetCompliancePolicy(id int) (*db.CompliancePolicy, error) {
	var policy db.CompliancePolicy
	err := d.Sql().SelectOne(&policy, "SELECT * FROM compliance_policies WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (d *SqlDb) GetCompliancePoliciesByProject(projectID int) ([]*db.CompliancePolicy, error) {
	var policies []*db.CompliancePolicy
	_, err := d.Sql().Select(&policies, "SELECT * FROM compliance_policies WHERE project_id = ? ORDER BY created DESC", projectID)
	return policies, err
}

func (d *SqlDb) UpdateCompliancePolicy(policy *db.CompliancePolicy) error {
	_, err := d.Sql().Update(policy)
	return err
}

func (d *SqlDb) DeleteCompliancePolicy(id int) error {
	_, err := d.Sql().Exec("DELETE FROM compliance_policies WHERE id = ?", id)
	return err
}

// PolicyAssignment methods

func (d *SqlDb) CreatePolicyAssignment(assignment *db.PolicyAssignment) error {
	return d.Sql().Insert(assignment)
}

func (d *SqlDb) GetPolicyAssignments(policyID int) ([]*db.PolicyAssignment, error) {
	var assignments []*db.PolicyAssignment
	_, err := d.Sql().Select(&assignments, "SELECT * FROM policy_assignments WHERE policy_id = ?", policyID)
	return assignments, err
}

func (d *SqlDb) DeletePolicyAssignments(policyID int) error {
	_, err := d.Sql().Exec("DELETE FROM policy_assignments WHERE policy_id = ?", policyID)
	return err
}

func (d *SqlDb) DeletePolicyAssignment(id int) error {
	_, err := d.Sql().Exec("DELETE FROM policy_assignments WHERE id = ?", id)
	return err
}

// ComplianceScan methods

func (d *SqlDb) CreateComplianceScan(scan *db.ComplianceScan) error {
	return d.Sql().Insert(scan)
}

func (d *SqlDb) GetComplianceScan(id int) (*db.ComplianceScan, error) {
	var scan db.ComplianceScan
	err := d.Sql().SelectOne(&scan, "SELECT * FROM compliance_scans WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &scan, nil
}

func (d *SqlDb) GetComplianceScansByProject(projectID int, limit, offset int) ([]*db.ComplianceScan, error) {
	var scans []*db.ComplianceScan
	query := "SELECT * FROM compliance_scans WHERE project_id = ? ORDER BY started DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
		if offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", offset)
		}
	}
	_, err := d.Sql().Select(&scans, query, projectID)
	return scans, err
}

func (d *SqlDb) GetComplianceScansByPolicy(policyID int) ([]*db.ComplianceScan, error) {
	var scans []*db.ComplianceScan
	_, err := d.Sql().Select(&scans, "SELECT * FROM compliance_scans WHERE policy_id = ? ORDER BY started DESC", policyID)
	return scans, err
}

func (d *SqlDb) UpdateComplianceScan(scan *db.ComplianceScan) error {
	_, err := d.Sql().Update(scan)
	return err
}

// ComplianceReport methods

func (d *SqlDb) CreateComplianceReport(report *db.ComplianceReport) error {
	return d.Sql().Insert(report)
}

func (d *SqlDb) GetComplianceReport(id int) (*db.ComplianceReport, error) {
	var report db.ComplianceReport
	err := d.Sql().SelectOne(&report, "SELECT * FROM compliance_reports WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (d *SqlDb) GetComplianceReportsByScan(scanID int) ([]*db.ComplianceReport, error) {
	var reports []*db.ComplianceReport
	_, err := d.Sql().Select(&reports, "SELECT * FROM compliance_reports WHERE scan_id = ? ORDER BY host", scanID)
	return reports, err
}

func (d *SqlDb) GetComplianceReportsByProject(projectID int, filters map[string]interface{}, limit, offset int) ([]*db.ComplianceReport, error) {
	var reports []*db.ComplianceReport
	
	// Build query with joins and filters
	query := squirrel.Select("cr.*").
		From("compliance_reports cr").
		Join("compliance_scans cs ON cr.scan_id = cs.id").
		Where(squirrel.Eq{"cs.project_id": projectID})
	
	// Add filters
	if policyID, ok := filters["policy_id"]; ok {
		query = query.Where(squirrel.Eq{"cs.policy_id": policyID})
	}
	if host, ok := filters["host"]; ok {
		query = query.Where(squirrel.Like{"cr.host": "%" + host.(string) + "%"})
	}
	if status, ok := filters["status"]; ok {
		query = query.Where(squirrel.Eq{"cr.result": status})
	}
	
	query = query.OrderBy("cr.created DESC")
	
	if limit > 0 {
		query = query.Limit(uint64(limit))
		if offset > 0 {
			query = query.Offset(uint64(offset))
		}
	}
	
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	
	_, err = d.Sql().Select(&reports, sql, args...)
	return reports, err
}

func (d *SqlDb) UpdateComplianceReport(report *db.ComplianceReport) error {
	_, err := d.Sql().Update(report)
	return err
}

// ComplianceRuleResult methods

func (d *SqlDb) CreateComplianceRuleResult(result *db.ComplianceRuleResult) error {
	return d.Sql().Insert(result)
}

func (d *SqlDb) GetComplianceRuleResultsByReport(reportID int) ([]*db.ComplianceRuleResult, error) {
	var results []*db.ComplianceRuleResult
	_, err := d.Sql().Select(&results, "SELECT * FROM compliance_rule_results WHERE report_id = ? ORDER BY rule_id", reportID)
	return results, err
}

func (d *SqlDb) DeleteComplianceRuleResultsByReport(reportID int) error {
	_, err := d.Sql().Exec("DELETE FROM compliance_rule_results WHERE report_id = ?", reportID)
	return err
}
