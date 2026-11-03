package projects

import (
	"fmt"
	"net/http"
	"time"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/schedules"
)

// SchedulesMiddleware ensures a template exists and loads it to the context
func SchedulesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		scheduleID, ok := helpers.GetIntParamOrAbort("schedule_id", w, r)
		if !ok { // not specified schedule_id
			return
		}

		schedule, err := helpers.Store(r).GetSchedule(project.ID, scheduleID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "schedule", schedule)
		next.ServeHTTP(w, r)
	})
}

func refreshSchedulePool(r *http.Request) {
	pool := helpers.GetFromContext(r, "schedule_pool").(schedules.SchedulePool)
	pool.Refresh()
}

// GetSchedule returns single template by ID
func GetSchedule(w http.ResponseWriter, r *http.Request) {
	schedule := helpers.GetFromContext(r, "schedule").(db.Schedule)
	helpers.WriteJSON(w, http.StatusOK, schedule)
}

func GetProjectSchedules(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	tplSchedules, err := helpers.Store(r).GetProjectSchedules(project.ID, false, false)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tplSchedules)
}
func GetTemplateSchedules(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	templateID, ok := helpers.GetIntParamOrAbort("template_id", w, r)
	if !ok {
		return
	}

	tplSchedules, err := helpers.Store(r).GetTemplateSchedules(project.ID, templateID, true)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tplSchedules)
}

func validateCronFormat(cronFormat string, w http.ResponseWriter) bool {
	err := schedules.ValidateCronFormat(cronFormat)
	if err == nil {
		return true
	}
	helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
		"error": "Cron: " + err.Error(),
	})
	return false
}

func validateSchedulePayload(schedule *db.Schedule, w http.ResponseWriter) bool {
	if schedule.Type == "" {
		schedule.Type = db.ScheduleTypeCron
	}

	switch schedule.Type {
	case db.ScheduleTypeRunAt:
		if schedule.RunAt == nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "run_at must be provided for run_at schedules",
			})
			return false
		}

		if schedule.RunAt.Before(time.Now()) {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "run_at must be in the future",
			})
			return false
		}

		schedule.CronFormat = ""
		return true
	case db.ScheduleTypeCron:
		schedule.RunAt = nil
		return validateCronFormat(schedule.CronFormat, w)
	default:
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid schedule type",
		})
		return false
	}
}

func ValidateScheduleCronFormat(w http.ResponseWriter, r *http.Request) {
	var schedule db.Schedule
	if !helpers.Bind(w, r, &schedule) {
		return
	}

	_ = validateCronFormat(schedule.CronFormat, w)
}

// AddSchedule creates a schedule.
func (c *ScheduleController) AddSchedule(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var schedule db.Schedule
	if !helpers.Bind(w, r, &schedule) {
		return
	}

	if !validateSchedulePayload(&schedule, w) {
		return
	}

	schedule.ProjectID = project.ID
	if !c.workflowTargetAllowed(w, r, schedule, nil) {
		return
	}
	schedule, err := helpers.Store(r).CreateSchedule(schedule)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventSchedule,
		ObjectID:    schedule.ID,
		Description: fmt.Sprintf("Schedule ID %d created", schedule.ID),
	})

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceScheduleCreate,
		Target:    audit.ResourceTarget(audit.TargetSchedule, schedule.ID, schedule.Name),
		ProjectID: project.ID,
		Metadata:  scheduleMetadata(schedule),
	})

	refreshSchedulePool(r)

	helpers.WriteJSON(w, http.StatusCreated, schedule)
}

// UpdateSchedule writes a schedule to an existing key in the database
func (c *ScheduleController) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	oldSchedule := helpers.GetFromContext(r, "schedule").(db.Schedule)

	var schedule db.Schedule
	if !helpers.Bind(w, r, &schedule) {
		return
	}

	// project ID and schedule ID in the body and the path must be the same

	if schedule.ID != oldSchedule.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "schedule id in URL and in body must be the same",
		})
		return
	}

	if schedule.ProjectID != oldSchedule.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "You can not move schedule to other project",
		})
		return
	}

	if !validateSchedulePayload(&schedule, w) {
		return
	}

	if !c.workflowTargetAllowed(w, r, schedule, &oldSchedule) {
		return
	}
	err := helpers.Store(r).UpdateSchedule(schedule)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldSchedule.ProjectID,
		ObjectType:  db.EventSchedule,
		ObjectID:    oldSchedule.ID,
		Description: fmt.Sprintf("Schedule ID %d updated", schedule.ID),
	})

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceScheduleUpdate,
		Target:    audit.ResourceTarget(audit.TargetSchedule, oldSchedule.ID, schedule.Name),
		ProjectID: oldSchedule.ProjectID,
		Metadata:  scheduleMetadata(schedule),
	})

	// A full update can switch the schedule on or off too.
	if oldSchedule.Active != schedule.Active {
		kind := audit.ResourceScheduleDeactivate
		if schedule.Active {
			kind = audit.ResourceScheduleActivate
		}
		helpers.Audit(r).Record(r.Context(), audit.Event{
			Kind:      kind,
			Target:    audit.ResourceTarget(audit.TargetSchedule, oldSchedule.ID, schedule.Name),
			ProjectID: oldSchedule.ProjectID,
			Metadata:  scheduleMetadata(schedule),
		})
	}

	refreshSchedulePool(r)

	w.WriteHeader(http.StatusNoContent)
}

func SetScheduleActive(w http.ResponseWriter, r *http.Request) {
	oldSchedule := helpers.GetFromContext(r, "schedule").(db.Schedule)

	var schedule struct {
		Active bool `json:"active"`
	}

	if !helpers.Bind(w, r, &schedule) {
		return
	}

	err := helpers.Store(r).SetScheduleActive(oldSchedule.ProjectID, oldSchedule.ID, schedule.Active)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldSchedule.ProjectID,
		ObjectType:  db.EventSchedule,
		ObjectID:    oldSchedule.ID,
		Description: fmt.Sprintf("Schedule ID %d updated", oldSchedule.ID),
	})

	if oldSchedule.Active != schedule.Active {
		kind := audit.ResourceScheduleDeactivate
		if schedule.Active {
			kind = audit.ResourceScheduleActivate
		}
		helpers.Audit(r).Record(r.Context(), audit.Event{
			Kind:      kind,
			Target:    audit.ResourceTarget(audit.TargetSchedule, oldSchedule.ID, oldSchedule.Name),
			ProjectID: oldSchedule.ProjectID,
			Metadata:  scheduleMetadata(oldSchedule),
		})
	}

	refreshSchedulePool(r)

	w.WriteHeader(http.StatusNoContent)
}

// RemoveSchedule deletes a schedule from the database
func RemoveSchedule(w http.ResponseWriter, r *http.Request) {
	schedule := helpers.GetFromContext(r, "schedule").(db.Schedule)

	err := helpers.Store(r).DeleteSchedule(schedule.ProjectID, schedule.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   schedule.ProjectID,
		ObjectType:  db.EventSchedule,
		ObjectID:    schedule.ID,
		Description: fmt.Sprintf("Schedule ID %d deleted", schedule.ID),
	})

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceScheduleDelete,
		Target:    audit.ResourceTarget(audit.TargetSchedule, schedule.ID, schedule.Name),
		ProjectID: schedule.ProjectID,
	})

	refreshSchedulePool(r)

	w.WriteHeader(http.StatusNoContent)
}

func scheduleMetadata(s db.Schedule) audit.ScheduleMetadata {
	m := audit.ScheduleMetadata{}
	if s.TemplateID != nil {
		m.TemplateID = *s.TemplateID
	}
	if s.WorkflowID != nil {
		m.WorkflowID = *s.WorkflowID
	}
	return m
}

// ScheduleController applies workflow availability to schedule mutations.
type ScheduleController struct{ pro proFeatureGate }

func NewScheduleController(subscription pro_interfaces.SubscriptionService) *ScheduleController {
	return &ScheduleController{pro: newProFeatureGate(subscription)}
}

func (c *ScheduleController) workflowTargetAllowed(w http.ResponseWriter, r *http.Request, s db.Schedule, old *db.Schedule) bool {
	if s.WorkflowID == nil || (old != nil && old.WorkflowID != nil && *old.WorkflowID == *s.WorkflowID) {
		return true
	}
	features, err := c.pro.features(helpers.UserFromContext(r))
	if err != nil {
		helpers.WriteErrorStatus(w, "Failed to read the subscription plan", http.StatusInternalServerError)
		return false
	}
	if !features.Workflows {
		helpers.WriteErrorStatus(w, "Workflow schedules are only available with Workflows enabled.", http.StatusForbidden)
		return false
	}
	return true
}
