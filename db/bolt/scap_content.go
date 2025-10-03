package bolt

import (
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/tz"
)

// SCAP Content operations for BoltDB

func (d *BoltDb) CreateScapContent(content *db.ScapContent) error {
	content.UploadedAt = tz.Now()
	return d.createObject(content.ProjectID, db.ScapContentProps, content)
}

func (d *BoltDb) GetScapContent(id int) (*db.ScapContent, error) {
	var content db.ScapContent
	err := d.getObjectByID(db.ScapContentProps, intObjectID(id), &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}

func (d *BoltDb) GetScapContentsByProject(projectID int) ([]*db.ScapContent, error) {
	var contents []*db.ScapContent
	err := d.getObjects(projectID, db.ScapContentProps, db.RetrieveQueryParams{}, nil, &contents)
	return contents, err
}

func (d *BoltDb) UpdateScapContent(content *db.ScapContent) error {
	return d.updateObject(content.ProjectID, db.ScapContentProps, content)
}

func (d *BoltDb) DeleteScapContent(id int) error {
	content, err := d.GetScapContent(id)
	if err != nil {
		return err
	}
	return d.deleteObject(content.ProjectID, db.ScapContentProps, intObjectID(id), nil)
}

// SCAP Profile operations

func (d *BoltDb) CreateScapProfile(profile *db.ScapProfile) error {
	profile.Created = tz.Now()
	return d.createObject(profile.ProjectID, db.ScapProfileProps, profile)
}

func (d *BoltDb) GetScapProfile(id int) (*db.ScapProfile, error) {
	var profile db.ScapProfile
	err := d.getObjectByID(db.ScapProfileProps, intObjectID(id), &profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (d *BoltDb) GetScapProfilesByContent(contentID int) ([]*db.ScapProfile, error) {
	var profiles []*db.ScapProfile
	err := d.getObjectsByForeignKey(db.ScapProfileProps, contentID, db.ScapContentProps, db.RetrieveQueryParams{}, &profiles)
	return profiles, err
}

func (d *BoltDb) DeleteScapProfile(id int) error {
	profile, err := d.GetScapProfile(id)
	if err != nil {
		return err
	}
	return d.deleteObject(profile.ProjectID, db.ScapProfileProps, intObjectID(id), nil)
}

// Compliance Policy operations

func (d *BoltDb) CreateCompliancePolicy(policy *db.CompliancePolicy) error {
	policy.Created = tz.Now()
	return d.createObject(policy.ProjectID, db.CompliancePolicyProps, policy)
}

func (d *BoltDb) GetCompliancePolicy(id int) (*db.CompliancePolicy, error) {
	var policy db.CompliancePolicy
	err := d.getObjectByID(db.CompliancePolicyProps, intObjectID(id), &policy)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (d *BoltDb) GetCompliancePoliciesByProject(projectID int) ([]*db.CompliancePolicy, error) {
	var policies []*db.CompliancePolicy
	err := d.getObjects(projectID, db.CompliancePolicyProps, db.RetrieveQueryParams{}, nil, &policies)
	return policies, err
}

func (d *BoltDb) UpdateCompliancePolicy(policy *db.CompliancePolicy) error {
	return d.updateObject(policy.ProjectID, db.CompliancePolicyProps, policy)
}

func (d *BoltDb) DeleteCompliancePolicy(id int) error {
	policy, err := d.GetCompliancePolicy(id)
	if err != nil {
		return err
	}
	return d.deleteObject(policy.ProjectID, db.CompliancePolicyProps, intObjectID(id), nil)
}

// Policy Assignment operations

func (d *BoltDb) CreatePolicyAssignment(assignment *db.PolicyAssignment) error {
	assignment.Created = tz.Now()
	return d.createObject(assignment.ProjectID, db.PolicyAssignmentProps, assignment)
}

func (d *BoltDb) GetPolicyAssignments(policyID int) ([]*db.PolicyAssignment, error) {
	var assignments []*db.PolicyAssignment
	err := d.getObjectsByForeignKey(db.PolicyAssignmentProps, policyID, db.CompliancePolicyProps, db.RetrieveQueryParams{}, &assignments)
	return assignments, err
}

func (d *BoltDb) DeletePolicyAssignment(id int) error {
	assignment, err := d.GetPolicyAssignment(id)
	if err != nil {
		return err
	}
	return d.deleteObject(assignment.ProjectID, db.PolicyAssignmentProps, intObjectID(id), nil)
}

func (d *BoltDb) GetPolicyAssignment(id int) (*db.PolicyAssignment, error) {
	var assignment db.PolicyAssignment
	err := d.getObjectByID(db.PolicyAssignmentProps, intObjectID(id), &assignment)
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// Compliance Scan operations

func (d *BoltDb) CreateComplianceScan(scan *db.ComplianceScan) error {
	scan.Created = tz.Now()
	return d.createObject(scan.ProjectID, db.ComplianceScanProps, scan)
}

func (d *BoltDb) GetComplianceScan(id int) (*db.ComplianceScan, error) {
	var scan db.ComplianceScan
	err := d.getObjectByID(db.ComplianceScanProps, intObjectID(id), &scan)
	if err != nil {
		return nil, err
	}
	return &scan, nil
}

func (d *BoltDb) GetComplianceScansByProject(projectID int, limit, offset int) ([]*db.ComplianceScan, error) {
	var scans []*db.ComplianceScan
	params := db.RetrieveQueryParams{
		Count:  limit,
		Offset: offset,
	}
	err := d.getObjects(projectID, db.ComplianceScanProps, params, nil, &scans)
	return scans, err
}

func (d *BoltDb) UpdateComplianceScan(scan *db.ComplianceScan) error {
	return d.updateObject(scan.ProjectID, db.ComplianceScanProps, scan)
}

// Compliance Report operations

func (d *BoltDb) CreateComplianceReport(report *db.ComplianceReport) error {
	report.Created = tz.Now()
	return d.createObject(report.ProjectID, db.ComplianceReportProps, report)
}

func (d *BoltDb) GetComplianceReport(id int) (*db.ComplianceReport, error) {
	var report db.ComplianceReport
	err := d.getObjectByID(db.ComplianceReportProps, intObjectID(id), &report)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (d *BoltDb) GetComplianceReportsByScan(scanID int) ([]*db.ComplianceReport, error) {
	var reports []*db.ComplianceReport
	err := d.getObjectsByForeignKey(db.ComplianceReportProps, scanID, db.ComplianceScanProps, db.RetrieveQueryParams{}, &reports)
	return reports, err
}

func (d *BoltDb) GetComplianceReportsByProject(projectID int, filters map[string]interface{}, limit, offset int) ([]*db.ComplianceReport, error) {
	var reports []*db.ComplianceReport
	params := db.RetrieveQueryParams{
		Count:  limit,
		Offset: offset,
	}
	err := d.getObjects(projectID, db.ComplianceReportProps, params, nil, &reports)
	return reports, err
}

// Compliance Rule Result operations

func (d *BoltDb) CreateComplianceRuleResult(result *db.ComplianceRuleResult) error {
	return d.createObject(result.ProjectID, db.ComplianceRuleResultProps, result)
}

func (d *BoltDb) GetComplianceRuleResultsByReport(reportID int) ([]*db.ComplianceRuleResult, error) {
	var results []*db.ComplianceRuleResult
	err := d.getObjectsByForeignKey(db.ComplianceRuleResultProps, reportID, db.ComplianceReportProps, db.RetrieveQueryParams{}, &results)
	return results, err
}
