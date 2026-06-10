package projects

import (
	"fmt"
	"net/http"
	"errors"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

func NotificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		notificationID, err := helpers.GetIntParam("notification_id", w, r)
		if err != nil {
			return
		}

		notification, err := helpers.Store(r).GetNotification(project.ID, notificationID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "notification", notification)
		next.ServeHTTP(w, r)
	})
}

func GetNotifications(w http.ResponseWriter, r *http.Request) {
	if notification := helpers.GetFromContext(r, "notification"); notification != nil {
		helpers.WriteJSON(w, http.StatusOK, notification.(db.Notification))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)
	
	notifications, err := helpers.Store(r).GetNotifications(project.ID, db.RetrieveQueryParams{})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, notifications)
}

func AddNotification(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var notification db.Notification
	if !helpers.Bind(w, r, &notification) {
		return
	}

	if notification.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	newNotification, err := helpers.Store(r).CreateNotification(notification)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventInventory,
		ObjectID:    newNotification.ID,
		Description: fmt.Sprintf("Notification %s created", notification.Name),
	})

	helpers.WriteJSON(w, http.StatusCreated, newNotification)
}

func UpdateNotification(w http.ResponseWriter, r *http.Request) {
	oldNotification := helpers.GetFromContext(r, "notification").(db.Notification)

	var notification db.Notification
	if !helpers.Bind(w, r, &notification) {
		return
	}

	if notification.ID != oldNotification.ID {
		helpers.WriteErrorStatus(w, "Notification ID in body and URL must be the same", http.StatusBadRequest)
		return
	}

	if notification.ProjectID != oldNotification.ProjectID {
		helpers.WriteErrorStatus(w, "Project ID in body and URL must be the same", http.StatusBadRequest)
		return
	}

	if err := helpers.Store(r).UpdateNotification(notification); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldNotification.ProjectID,
		ObjectType:  db.EventInventory,
		ObjectID:    oldNotification.ID,
		Description: fmt.Sprintf("Notification %s updated", notification.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}

func RemoveNotification(w http.ResponseWriter, r *http.Request) {
	notification := helpers.GetFromContext(r, "notification").(db.Notification)
	
	err := helpers.Store(r).DeleteNotification(notification.ProjectID, notification.ID)
	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Notification is in use by one or more templates",
			"inUse": true,
		})
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   notification.ProjectID,
		ObjectType:  db.EventNotification,
		ObjectID:    notification.ID,
		Description: fmt.Sprintf("Notification %s deleted", notification.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}

func GetNotificationRefs(w http.ResponseWriter, r *http.Request) {
	notification := helpers.GetFromContext(r, "notification").(db.Notification)
	
	refs, err := helpers.Store(r).GetNotificationRefs(notification.ProjectID, notification.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}
