package projects

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
	"net/http"
	"strconv"
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

// GetSchedule returns single template by ID
func GetSchedule(w http.ResponseWriter, r *http.Request) {
	schedule := context.Get(r, "schedule").(db.Schedule)
	helpers.WriteJSON(w, http.StatusOK, schedule)
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

	schedules, err := helpers.Store(r).GetTemplateSchedules(project.ID, templateID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, schedules)
}


// AddSchedule adds a template to the database
func AddSchedule(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var schedule db.Schedule
	if !helpers.Bind(w, r, &schedule) {
		return
	}

	schedule.ProjectID = project.ID
	schedule, err := helpers.Store(r).CreateSchedule(schedule)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)
	objType := "schedule"
	desc := "Schedule ID " + strconv.Itoa(schedule.ID) + " created"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &schedule.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

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

	err := helpers.Store(r).UpdateSchedule(schedule)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)

	desc := "Schedule ID " + strconv.Itoa(schedule.ID) + " updated"
	objType := "schedule"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &schedule.ProjectID,
		Description: &desc,
		ObjectID:    &schedule.ID,
		ObjectType:  &objType,
	})

	if err != nil {
		log.Error(err)
	}

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

	user := context.Get(r, "user").(*db.User)
	desc := "Schedule ID " + strconv.Itoa(schedule.ID) + " deleted"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &schedule.ProjectID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}