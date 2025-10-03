package compliance

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pkg/task_logger"
	"github.com/Digital-Data-Co/forge/pkg/tz"
	"github.com/Digital-Data-Co/forge/services/tasks"
	log "github.com/sirupsen/logrus"
)

// ComplianceJob represents a job for executing OpenSCAP compliance scans
type ComplianceJob struct {
	Task        db.Task
	ScanID      int
	PolicyID    int
	ContentFile string
	ProfileID   string
	Hosts       []string
	TaskPool    *tasks.TaskPool
}

// NewComplianceJob creates a new compliance job
func NewComplianceJob(task db.Task, scanID, policyID int, contentFile, profileID string, hosts []string, taskPool *tasks.TaskPool) *ComplianceJob {
	return &ComplianceJob{
		Task:        task,
		ScanID:      scanID,
		PolicyID:    policyID,
		ContentFile: contentFile,
		ProfileID:   profileID,
		Hosts:       hosts,
		TaskPool:    taskPool,
	}
}

// Run executes the compliance scan job
func (j *ComplianceJob) Run(username string, incomingVersion *string, alias string) error {
	// Create scan directory
	scanDir := filepath.Join("scans", fmt.Sprintf("scan_%d", j.ScanID))
	if err := os.MkdirAll(scanDir, 0755); err != nil {
		return fmt.Errorf("failed to create scan directory: %v", err)
	}

	// Create task log entry
	j.logTask("Starting compliance scan", fmt.Sprintf("Scan ID: %d, Policy ID: %d, Hosts: %s", j.ScanID, j.PolicyID, strings.Join(j.Hosts, ", ")))

	// Scan each host
	for i, host := range j.Hosts {
		j.logTask("Scanning host", fmt.Sprintf("Host %d/%d: %s", i+1, len(j.Hosts), host))

		if err := j.scanHost(host, scanDir); err != nil {
			j.logTask("Host scan failed", fmt.Sprintf("Host %s: %v", host, err))
			// Continue with other hosts
		} else {
			j.logTask("Host scan completed", fmt.Sprintf("Host %s: Success", host))
		}
	}

	j.logTask("Compliance scan completed", fmt.Sprintf("Scan ID: %d completed", j.ScanID))
	return nil
}

// scanHost performs a scan on a single host
func (j *ComplianceJob) scanHost(host, scanDir string) error {
	// Generate output file paths
	safeHost := strings.ReplaceAll(host, ":", "_")
	arfFile := filepath.Join(scanDir, fmt.Sprintf("%s.arf.xml", safeHost))
	reportFile := filepath.Join(scanDir, fmt.Sprintf("%s.report.html", safeHost))

	// Build oscap command
	cmd := exec.Command("oscap", "xccdf", "eval",
		"--profile", j.ProfileID,
		"--results-arf", arfFile,
		"--report", reportFile,
		j.ContentFile)

	// Set environment for remote execution if needed
	cmd.Env = append(os.Environ(), fmt.Sprintf("TARGET_HOST=%s", host))

	// Execute the scan
	output, err := cmd.CombinedOutput()
	if err != nil {
		j.logTask("Oscap command failed", fmt.Sprintf("Host: %s, Error: %v, Output: %s", host, err, string(output)))
		return fmt.Errorf("oscap scan failed for host %s: %v", host, err)
	}

	// Parse ARF results and create report
	if err := j.processScanResults(host, arfFile); err != nil {
		j.logTask("Failed to process results", fmt.Sprintf("Host: %s, Error: %v", host, err))
		return fmt.Errorf("failed to process scan results for host %s: %v", host, err)
	}

	return nil
}

// processScanResults processes the ARF results and creates a compliance report
func (j *ComplianceJob) processScanResults(host, arfFile string) error {
	// Parse ARF results
	parser := NewArfParser()
	summary, ruleResults, err := parser.ParseArfFile(arfFile)
	if err != nil {
		// Create error report if parsing fails
		return j.createErrorReport(host, fmt.Sprintf("Failed to parse ARF results: %v", err))
	}

	// Create success report
	return j.createSuccessReport(host, arfFile, summary, ruleResults)
}

// createSuccessReport creates a successful scan report
func (j *ComplianceJob) createSuccessReport(host, arfFile string, summary *ScanSummary, ruleResults []RuleResult) error {
	// Convert summary to JSON
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %v", err)
	}

	summaryStr := string(summaryJSON)

	// Create report
	report := &db.ComplianceReport{
		ScanID:  j.ScanID,
		Host:    host,
		Result:  strings.ToLower(summary.OverallResult),
		ArfPath: &arfFile,
		Summary: &summaryStr,
		Created: tz.Now(),
	}

	// Save report to database
	store := j.TaskPool.GetStore()
	if store == nil {
		return fmt.Errorf("store not available")
	}
	if err := store.CreateComplianceReport(report); err != nil {
		return fmt.Errorf("failed to save report: %v", err)
	}

	// Save rule results if available
	if len(ruleResults) > 0 {
		for _, rule := range ruleResults {
			dbRule := &db.ComplianceRuleResult{
				ReportID: report.ID,
				RuleID:   rule.ID,
				Severity: &rule.Severity,
				Result:   strings.ToLower(rule.Result),
			}

			if err := store.CreateComplianceRuleResult(dbRule); err != nil {
				log.Warnf("Failed to save rule result %s: %v", rule.ID, err)
			}
		}
	}

	return nil
}

// createErrorReport creates an error report for failed scans
func (j *ComplianceJob) createErrorReport(host, errorMsg string) error {
	summary := map[string]interface{}{
		"error":          errorMsg,
		"overall_result": "error",
	}

	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal error summary: %v", err)
	}

	summaryStr := string(summaryJSON)

	report := &db.ComplianceReport{
		ScanID:  j.ScanID,
		Host:    host,
		Result:  "error",
		Summary: &summaryStr,
		Created: tz.Now(),
	}

	store := j.TaskPool.GetStore()
	if store == nil {
		return fmt.Errorf("store not available")
	}
	return store.CreateComplianceReport(report)
}

// logTask logs a message to the task
func (j *ComplianceJob) logTask(message, details string) {
	// This would integrate with the task logging system
	// For now, we'll use standard logging
	log.WithFields(log.Fields{
		"task_id": j.Task.ID,
		"scan_id": j.ScanID,
		"message": message,
	}).Info(details)
}

// Kill stops the compliance job
func (j *ComplianceJob) Kill() {
	// Implementation for killing the job
	// This would need to integrate with the task system
}

// IsKilled checks if the job has been killed
func (j *ComplianceJob) IsKilled() bool {
	// Check if the task has been stopped
	task := j.TaskPool.GetTask(j.Task.ID)
	if task == nil {
		return true
	}
	return task.Task.Status == task_logger.TaskStoppingStatus
}

// CreateComplianceTask creates a new task for compliance scanning
func CreateComplianceTask(store db.Store, projectID, userID int, scanID, policyID int, contentFile, profileID string, hosts []string) (*db.Task, error) {
	// Create a compliance task
	task := &db.Task{
		ProjectID:  projectID,
		UserID:     &userID,
		TemplateID: 0, // No template for compliance tasks
		Status:     task_logger.TaskWaitingStatus,
		Created:    tz.Now(),
	}

	// Save task to database
	if _, err := store.CreateTask(*task, 0); err != nil {
		return nil, fmt.Errorf("failed to create task: %v", err)
	}

	return task, nil
}
