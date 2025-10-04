package bolt

import (
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/tz"
)

// SCAP Content operations for BoltDB

func (d *BoltDb) CreateScapContent(content *db.ScapContent) error {
	content.Created = tz.Now()
	_, err := d.createObject(content.ProjectID, db.ScapContentProps, content)
	return err
}

func (d *BoltDb) GetScapContent(id int) (*db.ScapContent, error) {
	var content db.ScapContent
	err := d.getObject(0, db.ScapContentProps, intObjectID(id), &content)
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
	_, err := d.createObject(0, db.ScapProfileProps, profile)
	return err
}

func (d *BoltDb) GetScapProfile(contentID int, profileID string) (*db.ScapProfile, error) {
	var profile db.ScapProfile
	// For BoltDB, we need to get all profiles and filter by contentID and profileID
	err := d.getObjects(0, db.ScapProfileProps, db.RetrieveQueryParams{}, func(obj any) bool {
		if p, ok := obj.(*db.ScapProfile); ok {
			return p.ContentID == contentID && p.ProfileID == profileID
		}
		return false
	}, &[]*db.ScapProfile{&profile})
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (d *BoltDb) GetScapProfilesByContent(contentID int) ([]*db.ScapProfile, error) {
	var profiles []*db.ScapProfile
	// For BoltDB, we need to get all profiles and filter by contentID
	err := d.getObjects(0, db.ScapProfileProps, db.RetrieveQueryParams{}, func(obj any) bool {
		if profile, ok := obj.(*db.ScapProfile); ok {
			return profile.ContentID == contentID
		}
		return false
	}, &profiles)
	return profiles, err
}

func (d *BoltDb) DeleteScapProfile(id int) error {
	return d.deleteObject(0, db.ScapProfileProps, intObjectID(id), nil)
}

// Compliance Policy operations

func (d *BoltDb) CreateCompliancePolicy(policy *db.CompliancePolicy) error {
	policy.Created = tz.Now()
	_, err := d.createObject(policy.ProjectID, db.CompliancePolicyProps, policy)
	return err
}

func (d *BoltDb) GetCompliancePolicy(id int) (*db.CompliancePolicy, error) {
	var policy db.CompliancePolicy
	err := d.getObject(0, db.CompliancePolicyProps, intObjectID(id), &policy)
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
	_, err := d.createObject(0, db.PolicyAssignmentProps, assignment)
	return err
}

func (d *BoltDb) GetPolicyAssignments(policyID int) ([]*db.PolicyAssignment, error) {
	var assignments []*db.PolicyAssignment
	// For BoltDB, we need to get all assignments and filter by policyID
	err := d.getObjects(0, db.PolicyAssignmentProps, db.RetrieveQueryParams{}, func(obj any) bool {
		if assignment, ok := obj.(*db.PolicyAssignment); ok {
			return assignment.PolicyID == policyID
		}
		return false
	}, &assignments)
	return assignments, err
}

func (d *BoltDb) DeletePolicyAssignment(id int) error {
	return d.deleteObject(0, db.PolicyAssignmentProps, intObjectID(id), nil)
}

func (d *BoltDb) GetPolicyAssignment(id int) (*db.PolicyAssignment, error) {
	var assignment db.PolicyAssignment
	err := d.getObject(0, db.PolicyAssignmentProps, intObjectID(id), &assignment)
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// Compliance Scan operations

func (d *BoltDb) CreateComplianceScan(scan *db.ComplianceScan) error {
	scan.Started = tz.Now()
	_, err := d.createObject(scan.ProjectID, db.ComplianceScanProps, scan)
	return err
}

func (d *BoltDb) GetComplianceScan(id int) (*db.ComplianceScan, error) {
	var scan db.ComplianceScan
	err := d.getObject(0, db.ComplianceScanProps, intObjectID(id), &scan)
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

func (d *BoltDb) GetComplianceScansByPolicy(policyID int) ([]*db.ComplianceScan, error) {
	var scans []*db.ComplianceScan
	// For BoltDB, we need to get all scans and filter by policyID
	err := d.getObjects(0, db.ComplianceScanProps, db.RetrieveQueryParams{}, func(obj any) bool {
		if scan, ok := obj.(*db.ComplianceScan); ok {
			return scan.PolicyID == policyID
		}
		return false
	}, &scans)
	return scans, err
}

func (d *BoltDb) UpdateComplianceScan(scan *db.ComplianceScan) error {
	return d.updateObject(scan.ProjectID, db.ComplianceScanProps, scan)
}

// Compliance Report operations

func (d *BoltDb) CreateComplianceReport(report *db.ComplianceReport) error {
	report.Created = tz.Now()
	_, err := d.createObject(0, db.ComplianceReportProps, report)
	return err
}

func (d *BoltDb) GetComplianceReport(id int) (*db.ComplianceReport, error) {
	var report db.ComplianceReport
	err := d.getObject(0, db.ComplianceReportProps, intObjectID(id), &report)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (d *BoltDb) GetComplianceReportsByScan(scanID int) ([]*db.ComplianceReport, error) {
	var reports []*db.ComplianceReport
	// For BoltDB, we need to get all reports and filter by scanID
	err := d.getObjects(0, db.ComplianceReportProps, db.RetrieveQueryParams{}, func(obj any) bool {
		if report, ok := obj.(*db.ComplianceReport); ok {
			return report.ScanID == scanID
		}
		return false
	}, &reports)
	return reports, err
}

func (d *BoltDb) UpdateComplianceReport(report *db.ComplianceReport) error {
	return d.updateObject(0, db.ComplianceReportProps, report)
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
	_, err := d.createObject(0, db.ComplianceRuleResultProps, result)
	return err
}

func (d *BoltDb) GetComplianceRuleResultsByReport(reportID int) ([]*db.ComplianceRuleResult, error) {
	var results []*db.ComplianceRuleResult
	// For BoltDB, we need to get all results and filter by reportID
	err := d.getObjects(0, db.ComplianceRuleResultProps, db.RetrieveQueryParams{}, func(obj any) bool {
		if result, ok := obj.(*db.ComplianceRuleResult); ok {
			return result.ReportID == reportID
		}
		return false
	}, &results)
	return results, err
}

func (d *BoltDb) DeleteComplianceRuleResultsByReport(reportID int) error {
	// Get all results for this report and delete them
	results, err := d.GetComplianceRuleResultsByReport(reportID)
	if err != nil {
		return err
	}

	for _, result := range results {
		err = d.deleteObject(0, db.ComplianceRuleResultProps, intObjectID(result.ID), nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *BoltDb) DeleteScapProfilesByContent(contentID int) error {
	// Get all profiles for this content and delete them
	profiles, err := d.GetScapProfilesByContent(contentID)
	if err != nil {
		return err
	}

	for _, profile := range profiles {
		err = d.DeleteScapProfile(profile.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *BoltDb) DeletePolicyAssignments(policyID int) error {
	// Get all assignments for this policy and delete them
	assignments, err := d.GetPolicyAssignments(policyID)
	if err != nil {
		return err
	}

	for _, assignment := range assignments {
		err = d.DeletePolicyAssignment(assignment.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
