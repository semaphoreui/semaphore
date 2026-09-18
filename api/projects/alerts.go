package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/tasks"
)

func AlertMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		alertID, ok := helpers.GetIntParamOrAbort("alert_id", w, r)
		if !ok {
			return
		}

		alert, err := helpers.Store(r).GetAlert(project.ID, alertID)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "alert", alert)
		next.ServeHTTP(w, r)
	})
}

func GetAlertDefaults(w http.ResponseWriter, r *http.Request) {
	bodies, err := tasks.DefaultAlertBodies()
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, bodies)
}

func GetAlerts(w http.ResponseWriter, r *http.Request) {
	if alert := helpers.GetFromContext(r, "alert"); alert != nil {
		helpers.WriteJSON(w, http.StatusOK, alert.(db.Alert))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)
	alerts, err := helpers.Store(r).GetAlerts(project.ID, helpers.QueryParams(r.URL))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, alerts)
}

func AddAlert(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	var alert db.Alert
	if !helpers.Bind(w, r, &alert) {
		return
	}

	if alert.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	alert.Normalize()
	if err := alert.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	newAlert, err := helpers.Store(r).CreateAlert(alert)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   newAlert.ProjectID,
		ObjectType:  db.EventAlert,
		ObjectID:    newAlert.ID,
		Description: fmt.Sprintf("Alert %s created", alert.Name),
	})

	helpers.WriteJSON(w, http.StatusCreated, newAlert)
}

func UpdateAlert(w http.ResponseWriter, r *http.Request) {
	old := helpers.GetFromContext(r, "alert").(db.Alert)
	var alert db.Alert
	if !helpers.Bind(w, r, &alert) {
		return
	}

	if alert.ID != old.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Alert ID in URL and in body must be the same",
		})
		return
	}

	alert.ProjectID = old.ProjectID
	if alert.Type == db.AlertTypeGotify && alert.Token == nil {
		alert.Token = old.Token
	}
	alert.Normalize()
	if alert.Type == db.AlertTypeGotify && alert.Token == nil {
		alert.Token = old.Token
	}
	if err := alert.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := helpers.Store(r).UpdateAlert(alert); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   old.ProjectID,
		ObjectType:  db.EventAlert,
		ObjectID:    old.ID,
		Description: fmt.Sprintf("Alert %s updated", alert.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}

func RemoveAlert(w http.ResponseWriter, r *http.Request) {
	alert := helpers.GetFromContext(r, "alert").(db.Alert)

	err := helpers.Store(r).DeleteAlert(alert.ProjectID, alert.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   alert.ProjectID,
		ObjectType:  db.EventAlert,
		ObjectID:    alert.ID,
		Description: fmt.Sprintf("Alert %s deleted", alert.Name),
	})

	w.WriteHeader(http.StatusNoContent)
}

func GetAlertRefs(w http.ResponseWriter, r *http.Request) {
	alert := helpers.GetFromContext(r, "alert").(db.Alert)
	refs, err := helpers.Store(r).GetAlertRefs(alert.ProjectID, alert.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	helpers.WriteJSON(w, http.StatusOK, refs)
}

func TestAlert(w http.ResponseWriter, r *http.Request) {
	alert := helpers.GetFromContext(r, "alert").(db.Alert)
	project := helpers.GetFromContext(r, "project").(db.Project)

	if err := tasks.SendAlertTest(project, alert, helpers.Store(r)); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
