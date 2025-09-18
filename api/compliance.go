package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// ComplianceDashboardData represents the data structure for compliance dashboard
type ComplianceDashboardData struct {
	Summary              ComplianceSummary              `json:"summary"`
	TaskCompliance       []TaskComplianceData           `json:"task_compliance"`
	UserActivity         []UserActivityData             `json:"user_activity"`
	ProjectCompliance    []ProjectComplianceData        `json:"project_compliance"`
	SecurityEvents       []SecurityEventData            `json:"security_events"`
	ComplianceTrends     ComplianceTrendsData           `json:"compliance_trends"`
	LastUpdated          time.Time                      `json:"last_updated"`
}

// ComplianceSummary provides high-level compliance metrics
type ComplianceSummary struct {
	TotalTasks           int     `json:"total_tasks"`
	SuccessfulTasks      int     `json:"successful_tasks"`
	FailedTasks          int     `json:"failed_tasks"`
	SuccessRate          float64 `json:"success_rate"`
	TotalUsers           int     `json:"total_users"`
	ActiveUsers          int     `json:"active_users"`
	TotalProjects        int     `json:"total_projects"`
	CompliantProjects    int     `json:"compliant_projects"`
	ComplianceRate       float64 `json:"compliance_rate"`
	SecurityIncidents    int     `json:"security_incidents"`
	LastAuditDate        *time.Time `json:"last_audit_date"`
}

// TaskComplianceData represents task-level compliance information
type TaskComplianceData struct {
	TaskID          int       `json:"task_id"`
	ProjectID       int       `json:"project_id"`
	ProjectName     string    `json:"project_name"`
	TemplateName    string    `json:"template_name"`
	Status          string    `json:"status"`
	Created         time.Time `json:"created"`
	Start           *time.Time `json:"start"`
	End             *time.Time `json:"end"`
	Duration        *int      `json:"duration_seconds"`
	UserID          *int      `json:"user_id"`
	Username        *string   `json:"username"`
	ComplianceScore int       `json:"compliance_score"`
	Issues          []string  `json:"issues"`
}

// UserActivityData represents user activity for compliance tracking
type UserActivityData struct {
	UserID       int       `json:"user_id"`
	Username     string    `json:"username"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	LastActivity *time.Time `json:"last_activity"`
	TotalTasks   int       `json:"total_tasks"`
	Admin        bool      `json:"admin"`
	External     bool      `json:"external"`
	Active       bool      `json:"active"`
}

// ProjectComplianceData represents project-level compliance information
type ProjectComplianceData struct {
	ProjectID        int       `json:"project_id"`
	ProjectName      string    `json:"project_name"`
	Created         time.Time `json:"created"`
	TotalTasks      int       `json:"total_tasks"`
	SuccessfulTasks int       `json:"successful_tasks"`
	FailedTasks     int       `json:"failed_tasks"`
	SuccessRate     float64   `json:"success_rate"`
	ComplianceScore int       `json:"compliance_score"`
	LastActivity    *time.Time `json:"last_activity"`
	TeamSize        int       `json:"team_size"`
	Issues          []string  `json:"issues"`
}

// SecurityEventData represents security-related events
type SecurityEventData struct {
	EventID      int       `json:"event_id"`
	EventType    string    `json:"event_type"`
	Description  string    `json:"description"`
	UserID       *int      `json:"user_id"`
	Username     *string   `json:"username"`
	ProjectID    *int      `json:"project_id"`
	ProjectName  *string   `json:"project_name"`
	Created      time.Time `json:"created"`
	Severity     string    `json:"severity"`
	Resolved     bool      `json:"resolved"`
}

// ComplianceTrendsData represents compliance trends over time
type ComplianceTrendsData struct {
	DailyTasks     []TrendDataPoint `json:"daily_tasks"`
	DailyUsers     []TrendDataPoint `json:"daily_users"`
	SuccessRates   []TrendDataPoint `json:"success_rates"`
	SecurityEvents []TrendDataPoint `json:"security_events"`
}

// TrendDataPoint represents a single data point in a trend
type TrendDataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
	Count int       `json:"count"`
}

// GetComplianceDashboard returns comprehensive compliance dashboard data
func GetComplianceDashboard(w http.ResponseWriter, r *http.Request) {
	user := helpers.GetFromContext(r, "user").(*db.User)
	
	// Check if user has admin privileges for compliance dashboard
	if !user.Admin {
		helpers.WriteErrorStatus(w, "Access denied: Admin privileges required", http.StatusForbidden)
		return
	}

	// Parse query parameters
	days := 30 // default to last 30 days
	if daysParam := r.URL.Query().Get("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	projectID := r.URL.Query().Get("project_id")
	
	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	store := helpers.Store(r)
	
	// Get summary data
	summary, err := getComplianceSummary(store, startDate, endDate, projectID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Failed to get compliance summary"})
		helpers.WriteError(w, err)
		return
	}

	// Get task compliance data
	taskCompliance, err := getTaskComplianceData(store, startDate, endDate, projectID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Failed to get task compliance data"})
		helpers.WriteError(w, err)
		return
	}

	// Get user activity data
	userActivity, err := getUserActivityData(store, startDate, endDate)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Failed to get user activity data"})
		helpers.WriteError(w, err)
		return
	}

	// Get project compliance data
	projectCompliance, err := getProjectComplianceData(store, startDate, endDate)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Failed to get project compliance data"})
		helpers.WriteError(w, err)
		return
	}

	// Get security events
	securityEvents, err := getSecurityEventsData(store, startDate, endDate, projectID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Failed to get security events data"})
		helpers.WriteError(w, err)
		return
	}

	// Get compliance trends
	trends, err := getComplianceTrends(store, startDate, endDate, projectID)
	if err != nil {
		util.LogErrorF(err, log.Fields{"error": "Failed to get compliance trends"})
		helpers.WriteError(w, err)
		return
	}

	dashboardData := ComplianceDashboardData{
		Summary:           summary,
		TaskCompliance:    taskCompliance,
		UserActivity:      userActivity,
		ProjectCompliance: projectCompliance,
		SecurityEvents:    securityEvents,
		ComplianceTrends:  trends,
		LastUpdated:       time.Now(),
	}

	helpers.WriteJSON(w, http.StatusOK, dashboardData)
}

// getComplianceSummary retrieves high-level compliance metrics
func getComplianceSummary(store db.Store, startDate, endDate time.Time, projectID string) (ComplianceSummary, error) {
	var summary ComplianceSummary
	
	// Get total tasks
	tasks, err := store.GetTasks(db.RetrieveQueryParams{
		SortBy: "created",
		SortInverted: true,
	})
	if err != nil {
		return summary, err
	}

	// Filter tasks by date range and project if specified
	var filteredTasks []db.Task
	for _, task := range tasks {
		if task.Created.After(startDate) && task.Created.Before(endDate) {
			if projectID == "" || strconv.Itoa(task.ProjectID) == projectID {
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	summary.TotalTasks = len(filteredTasks)
	
	// Count successful and failed tasks
	for _, task := range filteredTasks {
		if task.Status == "success" {
			summary.SuccessfulTasks++
		} else if task.Status == "failed" || task.Status == "stopped" {
			summary.FailedTasks++
		}
	}

	if summary.TotalTasks > 0 {
		summary.SuccessRate = float64(summary.SuccessfulTasks) / float64(summary.TotalTasks) * 100
	}

	// Get user statistics
	users, err := store.GetUsers(db.RetrieveQueryParams{})
	if err != nil {
		return summary, err
	}
	
	summary.TotalUsers = len(users)
	
	// Count active users (users with recent activity)
	activeUserCount := 0
	for _, user := range users {
		// Check if user has recent activity (last 7 days)
		userEvents, err := store.GetUserEvents(user.ID, db.RetrieveQueryParams{
			Count: 1,
		})
		if err == nil && len(userEvents) > 0 {
			if userEvents[0].Created.After(time.Now().AddDate(0, 0, -7)) {
				activeUserCount++
			}
		}
	}
	summary.ActiveUsers = activeUserCount

	// Get project statistics
	projects, err := store.GetAllProjects()
	if err != nil {
		return summary, err
	}
	
	summary.TotalProjects = len(projects)
	
	// Calculate compliant projects (projects with >80% success rate)
	compliantProjects := 0
	for _, project := range projects {
		projectTasks, err := store.GetTasks(db.RetrieveQueryParams{
			SortBy: "created",
			SortInverted: true,
		})
		if err != nil {
			continue
		}
		
		var projectTaskCount, projectSuccessCount int
		for _, task := range projectTasks {
			if task.ProjectID == project.ID && task.Created.After(startDate) && task.Created.Before(endDate) {
				projectTaskCount++
				if task.Status == "success" {
					projectSuccessCount++
				}
			}
		}
		
		if projectTaskCount > 0 {
			successRate := float64(projectSuccessCount) / float64(projectTaskCount) * 100
			if successRate >= 80 {
				compliantProjects++
			}
		}
	}
	
	summary.CompliantProjects = compliantProjects
	if summary.TotalProjects > 0 {
		summary.ComplianceRate = float64(compliantProjects) / float64(summary.TotalProjects) * 100
	}

	// Count security incidents (failed tasks with security implications)
	summary.SecurityIncidents = summary.FailedTasks // Simplified for now

	return summary, nil
}

// getTaskComplianceData retrieves task-level compliance information
func getTaskComplianceData(store db.Store, startDate, endDate time.Time, projectID string) ([]TaskComplianceData, error) {
	var taskCompliance []TaskComplianceData
	
	tasks, err := store.GetTasks(db.RetrieveQueryParams{
		SortBy: "created",
		SortInverted: true,
		Count: 100, // Limit to recent 100 tasks
	})
	if err != nil {
		return taskCompliance, err
	}

	for _, task := range tasks {
		if task.Created.After(startDate) && task.Created.Before(endDate) {
			if projectID != "" && strconv.Itoa(task.ProjectID) != projectID {
				continue
			}

			// Get project name
			project, err := store.GetProject(task.ProjectID)
			if err != nil {
				continue
			}

			// Get template name
			template, err := store.GetTemplate(task.ProjectID, task.TemplateID)
			if err != nil {
				continue
			}

			// Get username if available
			var username *string
			if task.UserID != nil {
				user, err := store.GetUser(*task.UserID)
				if err == nil {
					username = &user.Username
				}
			}

			// Calculate compliance score
			complianceScore := 100
			var issues []string

			if task.Status == "failed" || task.Status == "stopped" {
				complianceScore -= 50
				issues = append(issues, "Task execution failed")
			}

			if task.End != nil && task.Start != nil {
				duration := task.End.Sub(*task.Start)
				if duration.Minutes() > 60 { // Tasks taking more than 1 hour
					complianceScore -= 10
					issues = append(issues, "Task execution time exceeded threshold")
				}
			}

			// Calculate duration in seconds
			var duration *int
			if task.End != nil && task.Start != nil {
				dur := int(task.End.Sub(*task.Start).Seconds())
				duration = &dur
			}

			taskData := TaskComplianceData{
				TaskID:          task.ID,
				ProjectID:       task.ProjectID,
				ProjectName:     project.Name,
				TemplateName:    template.Name,
				Status:          string(task.Status),
				Created:         task.Created,
				Start:           task.Start,
				End:             task.End,
				Duration:        duration,
				UserID:          task.UserID,
				Username:        username,
				ComplianceScore: complianceScore,
				Issues:          issues,
			}

			taskCompliance = append(taskCompliance, taskData)
		}
	}

	return taskCompliance, nil
}

// getUserActivityData retrieves user activity information
func getUserActivityData(store db.Store, startDate, endDate time.Time) ([]UserActivityData, error) {
	var userActivity []UserActivityData
	
	users, err := store.GetUsers(db.RetrieveQueryParams{})
	if err != nil {
		return userActivity, err
	}

	for _, user := range users {
		// Get user's recent events
		userEvents, err := store.GetUserEvents(user.ID, db.RetrieveQueryParams{
			Count: 1,
		})
		if err != nil {
			continue
		}

		var lastActivity *time.Time
		if len(userEvents) > 0 {
			lastActivity = &userEvents[0].Created
		}

		// Count user's tasks in the date range
		tasks, err := store.GetTasks(db.RetrieveQueryParams{
			SortBy: "created",
			SortInverted: true,
		})
		if err != nil {
			continue
		}

		taskCount := 0
		for _, task := range tasks {
			if task.UserID != nil && *task.UserID == user.ID {
				if task.Created.After(startDate) && task.Created.Before(endDate) {
					taskCount++
				}
			}
		}

		// Determine if user is active (has activity in last 7 days)
		active := false
		if lastActivity != nil && lastActivity.After(time.Now().AddDate(0, 0, -7)) {
			active = true
		}

		userData := UserActivityData{
			UserID:       user.ID,
			Username:     user.Username,
			Name:         user.Name,
			Email:        user.Email,
			LastActivity: lastActivity,
			TotalTasks:   taskCount,
			Admin:        user.Admin,
			External:     user.External,
			Active:       active,
		}

		userActivity = append(userActivity, userData)
	}

	return userActivity, nil
}

// getProjectComplianceData retrieves project-level compliance information
func getProjectComplianceData(store db.Store, startDate, endDate time.Time) ([]ProjectComplianceData, error) {
	var projectCompliance []ProjectComplianceData
	
	projects, err := store.GetAllProjects()
	if err != nil {
		return projectCompliance, err
	}

	for _, project := range projects {
		// Get all tasks for this project
		tasks, err := store.GetTasks(db.RetrieveQueryParams{
			SortBy: "created",
			SortInverted: true,
		})
		if err != nil {
			continue
		}

		var projectTasks []db.Task
		var lastActivity *time.Time
		
		for _, task := range tasks {
			if task.ProjectID == project.ID && task.Created.After(startDate) && task.Created.Before(endDate) {
				projectTasks = append(projectTasks, task)
				if lastActivity == nil || task.Created.After(*lastActivity) {
					lastActivity = &task.Created
				}
			}
		}

		totalTasks := len(projectTasks)
		successfulTasks := 0
		failedTasks := 0

		for _, task := range projectTasks {
			if task.Status == "success" {
				successfulTasks++
			} else if task.Status == "failed" || task.Status == "stopped" {
				failedTasks++
			}
		}

		var successRate float64
		if totalTasks > 0 {
			successRate = float64(successfulTasks) / float64(totalTasks) * 100
		}

		// Calculate compliance score
		complianceScore := 100
		var issues []string

		if successRate < 80 {
			complianceScore -= 30
			issues = append(issues, "Low success rate")
		}

		if totalTasks == 0 {
			complianceScore -= 20
			issues = append(issues, "No recent activity")
		}

		// Get team size
		projectUsers, err := store.GetProjectUsers(project.ID)
		if err != nil {
			continue
		}
		teamSize := len(projectUsers)

		projectData := ProjectComplianceData{
			ProjectID:        project.ID,
			ProjectName:      project.Name,
			Created:          project.Created,
			TotalTasks:       totalTasks,
			SuccessfulTasks:  successfulTasks,
			FailedTasks:      failedTasks,
			SuccessRate:      successRate,
			ComplianceScore:  complianceScore,
			LastActivity:     lastActivity,
			TeamSize:         teamSize,
			Issues:           issues,
		}

		projectCompliance = append(projectCompliance, projectData)
	}

	return projectCompliance, nil
}

// getSecurityEventsData retrieves security-related events
func getSecurityEventsData(store db.Store, startDate, endDate time.Time, projectID string) ([]SecurityEventData, error) {
	var securityEvents []SecurityEventData
	
	// Get all events
	events, err := store.GetEvents(0, db.RetrieveQueryParams{
		SortBy: "created",
		SortInverted: true,
		Count: 200, // Limit to recent 200 events
	})
	if err != nil {
		return securityEvents, err
	}

	for _, event := range events {
		if event.Created.After(startDate) && event.Created.Before(endDate) {
			if projectID != "" && (event.ProjectID == nil || strconv.Itoa(*event.ProjectID) != projectID) {
				continue
			}

			// Determine if this is a security-related event
			isSecurityEvent := false
			severity := "low"
			
			if event.ObjectType != nil {
				switch *event.ObjectType {
				case "task":
					// Check if task failed (potential security concern)
					if event.Description != nil && 
						(strings.Contains(*event.Description, "failed") || 
						 strings.Contains(*event.Description, "stopped")) {
						isSecurityEvent = true
						severity = "medium"
					}
				case "user":
					// User-related events might be security relevant
					if event.Description != nil && 
						(strings.Contains(*event.Description, "login") || 
						 strings.Contains(*event.Description, "access") ||
						 strings.Contains(*event.Description, "permission")) {
						isSecurityEvent = true
						severity = "high"
					}
				}
			}

			if !isSecurityEvent {
				continue
			}

			// Get username if available
			var username *string
			if event.UserID != nil {
				user, err := store.GetUser(*event.UserID)
				if err == nil {
					username = &user.Username
				}
			}

			// Get project name if available
			var projectName *string
			if event.ProjectID != nil {
				project, err := store.GetProject(*event.ProjectID)
				if err == nil {
					projectName = &project.Name
				}
			}

			securityEvent := SecurityEventData{
				EventID:     event.ID,
				EventType:   string(*event.ObjectType),
				Description: *event.Description,
				UserID:      event.UserID,
				Username:    username,
				ProjectID:   event.ProjectID,
				ProjectName: projectName,
				Created:     event.Created,
				Severity:    severity,
				Resolved:    false, // Simplified for now
			}

			securityEvents = append(securityEvents, securityEvent)
		}
	}

	return securityEvents, nil
}

// getComplianceTrends retrieves compliance trends over time
func getComplianceTrends(store db.Store, startDate, endDate time.Time, projectID string) (ComplianceTrendsData, error) {
	var trends ComplianceTrendsData
	
	// Get daily task counts
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		dayStart := d
		dayEnd := d.AddDate(0, 0, 1)
		
		// Count tasks for this day
		tasks, err := store.GetTasks(db.RetrieveQueryParams{
			SortBy: "created",
			SortInverted: true,
		})
		if err != nil {
			continue
		}

		taskCount := 0
		successCount := 0
		for _, task := range tasks {
			if task.Created.After(dayStart) && task.Created.Before(dayEnd) {
				if projectID == "" || strconv.Itoa(task.ProjectID) == projectID {
					taskCount++
					if task.Status == "success" {
						successCount++
					}
				}
			}
		}

		successRate := 0.0
		if taskCount > 0 {
			successRate = float64(successCount) / float64(taskCount) * 100
		}

		trends.DailyTasks = append(trends.DailyTasks, TrendDataPoint{
			Date:  dayStart,
			Value: float64(taskCount),
			Count: taskCount,
		})

		trends.SuccessRates = append(trends.SuccessRates, TrendDataPoint{
			Date:  dayStart,
			Value: successRate,
			Count: taskCount,
		})

		// Count active users for this day
		users, err := store.GetUsers(db.RetrieveQueryParams{})
		if err == nil {
			activeUserCount := 0
			for _, user := range users {
				userEvents, err := store.GetUserEvents(user.ID, db.RetrieveQueryParams{
					Count: 1,
				})
				if err == nil && len(userEvents) > 0 {
					if userEvents[0].Created.After(dayStart) && userEvents[0].Created.Before(dayEnd) {
						activeUserCount++
					}
				}
			}

			trends.DailyUsers = append(trends.DailyUsers, TrendDataPoint{
				Date:  dayStart,
				Value: float64(activeUserCount),
				Count: activeUserCount,
			})
		}

		// Count security events for this day
		events, err := store.GetEvents(0, db.RetrieveQueryParams{
			SortBy: "created",
			SortInverted: true,
		})
		if err == nil {
			securityEventCount := 0
			for _, event := range events {
				if event.Created.After(dayStart) && event.Created.Before(dayEnd) {
					if event.ObjectType != nil && 
						(*event.ObjectType == "task" || *event.ObjectType == "user") {
						securityEventCount++
					}
				}
			}

			trends.SecurityEvents = append(trends.SecurityEvents, TrendDataPoint{
				Date:  dayStart,
				Value: float64(securityEventCount),
				Count: securityEventCount,
			})
		}
	}

	return trends, nil
}
