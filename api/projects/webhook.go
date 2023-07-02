package projects

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"

	"github.com/gorilla/context"
)

func WebhookMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid webhook ID",
			});
		}

		project := context.Get(r, "project").(db.Project)
		webhook, err := helpers.Store(r).GetWebhook(project.ID, webhook_id)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "webhook", webhook)
		next.ServeHTTP(w, r)
	})
}

func GetWebhook(w http.ResponseWriter, r *http.Request) {
	webhook := context.Get(r, "webhook").(db.Webhook)
	helpers.WriteJSON(w, http.StatusOK, webhook)
}


func GetWebhooks(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	webhooks, err := helpers.Store(r).GetWebhooks(project.ID, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, webhooks)
}

func GetWebhookRefs (w http.ResponseWriter, r *http.Request) {
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Webhook ID",
		})
	}

	project := context.Get(r, "project").(db.Project)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	refs, err := helpers.Store(r).GetWebhookRefs(project.ID, webhook_id)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}


func AddWebhook(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	var webhook db.Webhook

	if !helpers.Bind(w, r, &webhook) {
		return
	}

	if webhook.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string {
			"error": "Project ID in body and URL must be the same",
		})
		return
	}
	err := webhook.Validate()
	if  err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string {
			"error": err.Error(),
	    })
		return
	}

	newWebhook, errWebhook := helpers.Store(r).CreateWebhook(webhook)

	if errWebhook != nil {
		helpers.WriteError(w, errWebhook)
		return
	}

	user := context.Get(r, "user").(*db.User)

	objType := db.EventWebhook
	desc := "Webhook " + webhook.Name + " created"
	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &newWebhook.ID,
		Description: &desc,
	})


	w.WriteHeader(http.StatusNoContent)
}


func UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	oldWebhook := context.Get(r, "webhook").(db.Webhook)
	var webhook db.Webhook

	if !helpers.Bind(w, r, &webhook) {
		return
	}

	if webhook.ID != oldWebhook.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Webhook ID in body and URL must be the same",
		})
		return
	}

	if webhook.ProjectID != oldWebhook.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	err := helpers.Store(r).UpdateWebhook(webhook)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := context.Get(r, "user").(*db.User)

	desc := "Webhook (" + webhook.Name + ") updated"
	objType := db.EventWebhook

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &webhook.ProjectID,
		Description: &desc,
		ObjectID:    &webhook.ID,
		ObjectType:  &objType,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}


func DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	webhook := context.Get(r, "webhook").(db.Webhook)
	project := context.Get(r, "project").(db.Project)

	err := helpers.Store(r).DeleteWebhook(project.ID, webhook.ID)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Webhook failed to be deleted",
		})
	}

	user := context.Get(r, "user").(*db.User)

	desc := "Webhook " + webhook.Name + " deleted"

	_, err = helpers.Store(r).CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &project.ID,
		Description: &desc,
	})

	if err != nil {
		log.Error(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
