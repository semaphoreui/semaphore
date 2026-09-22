package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/alerting"
)

// AlertController exposes project alerts: named notification destinations
// that templates and schedules bind to.
type AlertController struct {
	alertService alerting.Service
}

func NewAlertController(alertService alerting.Service) *AlertController {
	return &AlertController{alertService: alertService}
}

// AlertMiddleware loads the alert addressed by {alert_id} within the current
// project into the request context.
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

// GetChannels describes the supported channels: fields, default events,
// default message template and whether the server has them configured.
func (c *AlertController) GetChannels(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	helpers.WriteJSON(w, http.StatusOK, c.alertService.Channels(project))
}

func (c *AlertController) GetAlerts(w http.ResponseWriter, r *http.Request) {
	if alert := helpers.GetFromContext(r, "alert"); alert != nil {
		helpers.WriteJSON(w, http.StatusOK, alert.(db.Alert))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)

	alerts, err := helpers.Store(r).GetAlerts(project.ID, helpers.QueryParamsForProps(r.URL, db.AlertProps))
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, alerts)
}

func (c *AlertController) AddAlert(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var alert db.Alert
	if !helpers.Bind(w, r, &alert) {
		return
	}

	if alert.ProjectID != project.ID {
		helpers.WriteErrorStatus(w, "Project ID in body and URL must be the same", http.StatusBadRequest)
		return
	}

	if err := c.alertService.ValidateAlert(&alert); err != nil {
		helpers.WriteError(w, err)
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
		Description: fmt.Sprintf("Alert %s created", newAlert.Name),
	})

	helpers.WriteJSON(w, http.StatusCreated, newAlert)
}

func (c *AlertController) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	old := helpers.GetFromContext(r, "alert").(db.Alert)

	var alert db.Alert
	if !helpers.Bind(w, r, &alert) {
		return
	}

	if alert.ID != old.ID {
		helpers.WriteErrorStatus(w, "Alert ID in URL and in body must be the same", http.StatusBadRequest)
		return
	}
	alert.ProjectID = old.ProjectID

	if err := c.alertService.ValidateAlert(&alert); err != nil {
		helpers.WriteError(w, err)
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

func (c *AlertController) RemoveAlert(w http.ResponseWriter, r *http.Request) {
	alert := helpers.GetFromContext(r, "alert").(db.Alert)

	if err := helpers.Store(r).DeleteAlert(alert.ProjectID, alert.ID); err != nil {
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

func (c *AlertController) GetAlertRefs(w http.ResponseWriter, r *http.Request) {
	alert := helpers.GetFromContext(r, "alert").(db.Alert)

	refs, err := helpers.Store(r).GetAlertRefs(alert.ProjectID, alert.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// TestAlert sends a test message to one alert.
func (c *AlertController) TestAlert(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	alert := helpers.GetFromContext(r, "alert").(db.Alert)

	if err := c.alertService.SendTest(r.Context(), project, alert); err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SendTestNotification sends a test message through every destination the
// project uses by default: the server channels (when the project allows
// them) and every enabled project alert. 409 means there is nothing to send.
func (c *AlertController) SendTestNotification(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	sent, err := c.alertService.SendProjectTest(r.Context(), project)
	if sent == 0 && err == nil {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
