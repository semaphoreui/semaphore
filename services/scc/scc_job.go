package scc

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/task_logger"
	"github.com/Digital-Data-Co/forge/pkg/tz"
	log "github.com/sirupsen/logrus"
)

// SCCJob represents an SCC compliance scan job
type SCCJob struct {
	TaskID     int
	Store      db.Store
	SCCService *SCCService
	ProjectID  int
	PolicyID   int
	Host       string
	Benchmark  string
	Profile    string
	UserID     int
}

// NewSCCJob creates a new SCC compliance scan job
func NewSCCJob(taskID int, store db.Store, sccService *SCCService, projectID, policyID int, host, benchmark, profile string, userID int) *SCCJob {
	return &SCCJob{
		TaskID:     taskID,
		Store:      store,
		SCCService: sccService,
		ProjectID:  projectID,
		PolicyID:   policyID,
		Host:       host,
		Benchmark:  benchmark,
		Profile:    profile,
		UserID:     userID,
	}
}

// Run executes the SCC compliance scan
func (j *SCCJob) Run(ctx context.Context) error {
	// Create a compliance scan record
	scan := &db.ComplianceScan{
		ProjectID:   j.ProjectID,
		PolicyID:    j.PolicyID,
		InitiatedBy: j.UserID,
		Started:     tz.Now(),
		Status:      "running",
	}

	err := j.Store.CreateComplianceScan(scan)
	if err != nil {
		return fmt.Errorf("failed to create compliance scan: %w", err)
	}

	// Update task with scan ID
	task, err := j.Store.GetTask(j.TaskID, j.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Playbook = fmt.Sprintf("SCC Scan - %s (%s)", j.Host, j.Profile)
	err = j.Store.UpdateTask(task)
	if err != nil {
		log.Errorf("Failed to update task name: %v", err)
	}

	// Run the SCC scan
	result, err := j.SCCService.RunSCCScan(ctx, j.Host, j.Benchmark, j.Profile)
	if err != nil {
		// Update scan status to failed
		scan.Status = "failed"
		ended := tz.Now()
		scan.Ended = &ended
		j.Store.UpdateComplianceScan(scan)

		return fmt.Errorf("SCC scan failed: %w", err)
	}

	// Update scan status
	scan.Status = "completed"
	ended := tz.Now()
	scan.Ended = &ended
	err = j.Store.UpdateComplianceScan(scan)
	if err != nil {
		log.Errorf("Failed to update compliance scan: %v", err)
	}

	// Create compliance report
	report := &db.ComplianceReport{
		ScanID:  scan.ID,
		Host:    j.Host,
		Result:  result.Status,
		Created: tz.Now(),
	}

	// Add summary information
	summary := map[string]interface{}{
		"score":          result.Score,
		"total_rules":    result.TotalRules,
		"passed_rules":   result.PassedRules,
		"failed_rules":   result.FailedRules,
		"not_applicable": result.NotApplicable,
		"not_checked":    result.NotChecked,
		"duration":       result.Duration.String(),
		"benchmark":      result.Benchmark,
		"profile":        result.Profile,
	}

	report.SetSummary(summary)
	report.ArfPath = &result.ReportPath

	err = j.Store.CreateComplianceReport(report)
	if err != nil {
		log.Errorf("Failed to create compliance report: %v", err)
		return fmt.Errorf("failed to create compliance report: %w", err)
	}

	// Store individual rule results
	for _, ruleResult := range result.Results {
		dbRuleResult := &db.ComplianceRuleResult{
			ReportID:    report.ID,
			RuleID:      ruleResult.RuleID,
			Severity:    &ruleResult.Severity,
			Result:      ruleResult.Status,
			Ident:       &ruleResult.Title,
			Description: &ruleResult.Description,
		}

		err = j.Store.CreateComplianceRuleResult(dbRuleResult)
		if err != nil {
			log.Errorf("Failed to create rule result: %v", err)
		}
	}

	// Create task files for the scan results
	j.createTaskFiles(result)

	log.Infof("SCC scan completed successfully for host %s", j.Host)
	return nil
}

// Kill terminates the SCC scan (not supported by SCC, but we can update status)
func (j *SCCJob) Kill() error {
	// SCC scans can't be easily killed once started, but we can update the status
	log.Warnf("Kill requested for SCC scan on host %s - scan will continue to completion", j.Host)
	return nil
}

// createTaskFiles creates task files for the scan results
func (j *SCCJob) createTaskFiles(result *SCCScanResult) {
	// Create task files for various outputs
	taskFiles := []struct {
		name     string
		path     string
		mimeType string
	}{
		{"scan_results.json", filepath.Join(result.ReportPath, "results.json"), "application/json"},
		{"scan_report.html", filepath.Join(result.ReportPath, "report.html"), "text/html"},
		{"scan_results.xml", filepath.Join(result.ReportPath, "results.xml"), "application/xml"},
		{"scan_summary.txt", filepath.Join(result.ReportPath, "summary.txt"), "text/plain"},
	}

	for _, tf := range taskFiles {
		taskFile := &db.TaskFile{
			TaskID:       j.TaskID,
			ProjectID:    j.ProjectID,
			Filename:     tf.name,
			OriginalPath: tf.path,
			FileSize:     0, // Will be updated when file is actually created
			MimeType:     tf.mimeType,
		}

		_, err := j.Store.CreateTaskFile(*taskFile)
		if err != nil {
			log.Errorf("Failed to create task file %s: %v", tf.name, err)
		}
	}
}

// CreateSCCTask creates a new SCC compliance scan task
func CreateSCCTask(store db.Store, sccService *SCCService, projectID, policyID int, host, benchmark, profile string, userID int) (*db.Task, error) {
	// Create the task
	task := &db.Task{
		ProjectID: projectID,
		Status:    task_logger.TaskWaitingStatus,
		Created:   tz.Now(),
		UserID:    &userID,
	}

	createdTask, err := store.CreateTask(*task, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create SCC task: %w", err)
	}

	// Create the SCC job
	sccJob := NewSCCJob(createdTask.ID, store, sccService, projectID, policyID, host, benchmark, profile, userID)

	// TODO: Register the job with the task pool when task pool integration is available
	_ = sccJob // Suppress unused variable warning

	return &createdTask, nil
}

// SCCScanRequest represents a request to run an SCC scan
type SCCScanRequest struct {
	ProjectID int    `json:"project_id"`
	PolicyID  int    `json:"policy_id"`
	Host      string `json:"host"`
	Benchmark string `json:"benchmark"`
	Profile   string `json:"profile"`
	UserID    int    `json:"user_id"`
}

// ValidateSCCScanRequest validates an SCC scan request
func ValidateSCCScanRequest(req *SCCScanRequest) error {
	if req.ProjectID <= 0 {
		return fmt.Errorf("invalid project ID")
	}
	if req.PolicyID <= 0 {
		return fmt.Errorf("invalid policy ID")
	}
	if req.Host == "" {
		return fmt.Errorf("host is required")
	}
	if req.Benchmark == "" {
		return fmt.Errorf("benchmark is required")
	}
	if req.Profile == "" {
		return fmt.Errorf("profile is required")
	}
	if req.UserID <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	return nil
}

// GetSCCScanHistory returns the scan history for a project
func GetSCCScanHistory(store db.Store, projectID int, limit, offset int) ([]*db.ComplianceScan, error) {
	return store.GetComplianceScansByProject(projectID, limit, offset)
}

// GetSCCScanResults returns the results for a specific scan
func GetSCCScanResults(store db.Store, scanID int) ([]*db.ComplianceReport, error) {
	return store.GetComplianceReportsByScan(scanID)
}

// GetSCCRuleResults returns the rule results for a specific report
func GetSCCRuleResults(store db.Store, reportID int) ([]*db.ComplianceRuleResult, error) {
	return store.GetComplianceRuleResultsByReport(reportID)
}
