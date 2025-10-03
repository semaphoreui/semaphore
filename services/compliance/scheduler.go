package compliance

import (
	"fmt"

	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/services/schedules"
	log "github.com/sirupsen/logrus"
)

// ComplianceScheduleRunner handles scheduled compliance scans
type ComplianceScheduleRunner struct {
	projectID  int
	policyID   int
	scannerSvc *ScannerService
	policySvc  *PolicyService
}

// NewComplianceScheduleRunner creates a new compliance schedule runner
func NewComplianceScheduleRunner(projectID, policyID int, scannerSvc *ScannerService, policySvc *PolicyService) *ComplianceScheduleRunner {
	return &ComplianceScheduleRunner{
		projectID:  projectID,
		policyID:   policyID,
		scannerSvc: scannerSvc,
		policySvc:  policySvc,
	}
}

// Run executes the scheduled compliance scan
func (r *ComplianceScheduleRunner) Run() {
	log.WithFields(log.Fields{
		"project_id": r.projectID,
		"policy_id":  r.policyID,
	}).Info("Running scheduled compliance scan")

	// Get the policy
	policy, err := r.policySvc.GetPolicy(r.policyID)
	if err != nil {
		log.WithError(err).Error("Failed to get policy for scheduled scan")
		return
	}

	// Verify policy belongs to project
	if policy.ProjectID != r.projectID {
		log.Error("Policy does not belong to project")
		return
	}

	// Check if policy is still active (has assignments)
	assignments, err := r.policySvc.GetPolicyAssignments(r.policyID)
	if err != nil {
		log.WithError(err).Error("Failed to get policy assignments")
		return
	}

	if len(assignments) == 0 {
		log.Info("Policy has no assignments, skipping scheduled scan")
		return
	}

	// Initiate scan (using system user ID 0 for scheduled scans)
	_, err = r.scannerSvc.ScanPolicy(r.projectID, 0, r.policyID)
	if err != nil {
		log.WithError(err).Error("Failed to initiate scheduled compliance scan")
		return
	}

	log.WithFields(log.Fields{
		"project_id": r.projectID,
		"policy_id":  r.policyID,
	}).Info("Scheduled compliance scan initiated successfully")
}

// CreateComplianceSchedule creates a schedule for a compliance policy
func CreateComplianceSchedule(store db.Store, projectID, policyID int, cronFormat, name string) (*db.Schedule, error) {
	// Create a schedule entry
	schedule := &db.Schedule{
		ProjectID:  projectID,
		TemplateID: 0, // No template for compliance schedules
		CronFormat: cronFormat,
		Name:       name,
		Active:     true,
	}

	// Save schedule to database
	newSchedule, err := store.CreateSchedule(*schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to create schedule: %v", err)
	}

	return &newSchedule, nil
}

// UpdateComplianceSchedule updates a compliance policy's schedule
func UpdateComplianceSchedule(store db.Store, policy *db.CompliancePolicy, scheduleID *int, cronFormat, name string) error {
	if scheduleID == nil {
		// Create new schedule
		schedule, err := CreateComplianceSchedule(store, policy.ProjectID, policy.ID, cronFormat, name)
		if err != nil {
			return err
		}

		// Update policy with new schedule ID
		policy.ScheduleID = &schedule.ID
		return store.UpdateCompliancePolicy(policy)
	} else {
		// Update existing schedule
		schedule, err := store.GetSchedule(policy.ProjectID, *scheduleID)
		if err != nil {
			return fmt.Errorf("failed to get schedule: %v", err)
		}

		schedule.CronFormat = cronFormat
		schedule.Name = name
		schedule.Active = true

		return store.UpdateSchedule(schedule)
	}
}

// DeleteComplianceSchedule deletes a compliance policy's schedule
func DeleteComplianceSchedule(store db.Store, policy *db.CompliancePolicy) error {
	if policy.ScheduleID == nil {
		return nil // No schedule to delete
	}

	// Delete the schedule
	err := store.DeleteSchedule(policy.ProjectID, *policy.ScheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %v", err)
	}

	// Update policy to remove schedule reference
	policy.ScheduleID = nil
	return store.UpdateCompliancePolicy(policy)
}

// ComplianceScheduleManager manages compliance schedules
type ComplianceScheduleManager struct {
	store        db.Store
	scannerSvc   *ScannerService
	policySvc    *PolicyService
	schedulePool *schedules.SchedulePool
}

// NewComplianceScheduleManager creates a new compliance schedule manager
func NewComplianceScheduleManager(store db.Store, scannerSvc *ScannerService, policySvc *PolicyService, schedulePool *schedules.SchedulePool) *ComplianceScheduleManager {
	return &ComplianceScheduleManager{
		store:        store,
		scannerSvc:   scannerSvc,
		policySvc:    policySvc,
		schedulePool: schedulePool,
	}
}

// RefreshComplianceSchedules refreshes all compliance schedules
func (m *ComplianceScheduleManager) RefreshComplianceSchedules() error {
	// Get all active compliance policies with schedules
	// This would need to be implemented as a query that joins policies with schedules
	// For now, we'll get all policies and check which ones have schedules

	// Get all projects
	projects, err := m.store.GetAllProjects()
	if err != nil {
		return fmt.Errorf("failed to get projects: %v", err)
	}

	for _, project := range projects {
		policies, err := m.policySvc.GetPoliciesByProject(project.ID)
		if err != nil {
			log.WithError(err).WithField("project_id", project.ID).Warn("Failed to get policies for project")
			continue
		}

		for _, policy := range policies {
			if policy.ScheduleID != nil {
				// Get the schedule
				schedule, err := m.store.GetSchedule(policy.ProjectID, *policy.ScheduleID)
				if err != nil {
					log.WithError(err).WithFields(log.Fields{
						"project_id":  policy.ProjectID,
						"policy_id":   policy.ID,
						"schedule_id": *policy.ScheduleID,
					}).Warn("Failed to get schedule for policy")
					continue
				}

				if schedule.Active {
					// Create compliance schedule runner
					runner := NewComplianceScheduleRunner(policy.ProjectID, policy.ID, m.scannerSvc, m.policySvc)

					// Add to schedule pool
					_, err = m.schedulePool.AddComplianceRunner(runner, schedule.CronFormat)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"project_id":  policy.ProjectID,
							"policy_id":   policy.ID,
							"schedule_id": *policy.ScheduleID,
						}).Warn("Failed to add compliance schedule runner")
					}
				}
			}
		}
	}

	return nil
}

// GetComplianceScheduleStatus returns the status of a compliance policy's schedule
func (m *ComplianceScheduleManager) GetComplianceScheduleStatus(policyID int) (map[string]interface{}, error) {
	policy, err := m.policySvc.GetPolicy(policyID)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"policy_id":    policy.ID,
		"has_schedule": policy.ScheduleID != nil,
	}

	if policy.ScheduleID != nil {
		schedule, err := m.store.GetSchedule(policy.ProjectID, *policy.ScheduleID)
		if err != nil {
			result["schedule_error"] = err.Error()
		} else {
			result["schedule"] = map[string]interface{}{
				"id":          schedule.ID,
				"name":        schedule.Name,
				"cron_format": schedule.CronFormat,
				"active":      schedule.Active,
			}
		}
	}

	return result, nil
}
