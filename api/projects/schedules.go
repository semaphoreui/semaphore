package projects

import (
	"fmt"
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/services/schedules"
	"github.com/gorilla/context"
)

// SchedulesMiddleware ensures a template exists and loads it to the context
func SchedulesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		scheduleID, err := helpers.GetIntParam("schedule_id", w, r)
		if err != nil { // not specified schedule_id
			return
		}

		schedule, err := helpers.Store(r).GetSchedule(project.ID, scheduleID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "schedule", schedule)
		next.ServeHTTP(w, r)
	})
}

func refreshSchedulePool(r *http.Request) {
	pool := context.Get(r, "schedule_pool").(schedules.SchedulePool)
	pool.Refresh()
}

// GetSchedule returns single template by ID
func GetSchedule(w http.ResponseWriter, r *http.Request) {
	schedule := context.Get(r, "schedule").(db.Schedule)
	helpers.WriteJSON(w, http.StatusOK, schedule)
}

func GetProjectSchedules(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	tplSchedules, err := helpers.Store(r).GetProjectSchedules(project.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, tplSchedules)
}
func GetTemplateSchedules(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	templateID, err := helpers.GetIntParam("template_id", w, r)
	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "template_id must be provided",
		})
		return
	}

	tplSchedules, err := helpers.Store(r).GetTemplateSchedules(project.ID, templateID)
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

func ValidateScheduleCronFormat(w http.ResponseWriter, r *http.Request) {
	var schedule db.Schedule
	if !helpers.Bind(w, r, &schedule) {
		return
	}

	_ = validateCronFormat(schedule.CronFormat, w)
}

// AddSchedule adds a template to the database
func AddSchedule(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var schedule db.Schedule
	if !helpers.Bind(w, r, &schedule) {
		return
	}

	if !validateCronFormat(schedule.CronFormat, w) {
		return
	}

	schedule.ProjectID = project.ID
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

	refreshSchedulePool(r)

	helpers.WriteJSON(w, http.StatusCreated, schedule)
}

// UpdateSchedule writes a schedule to an existing key in the database
func UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	oldSchedule := context.Get(r, "schedule").(db.Schedule)

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

	if !validateCronFormat(schedule.CronFormat, w) {
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

	refreshSchedulePool(r)

	w.WriteHeader(http.StatusNoContent)
}

func SetScheduleActive(w http.ResponseWriter, r *http.Request) {
	oldSchedule := context.Get(r, "schedule").(db.Schedule)

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

	refreshSchedulePool(r)

	w.WriteHeader(http.StatusNoContent)
}

// RemoveSchedule deletes a schedule from the database
func RemoveSchedule(w http.ResponseWriter, r *http.Request) {
	schedule := context.Get(r, "schedule").(db.Schedule)

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

	refreshSchedulePool(r)

	w.WriteHeader(http.StatusNoContent)
}
