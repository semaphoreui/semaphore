package projects

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/gorilla/context"
)

func WebhookMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhook_id, err := helpers.GetIntParam("webhook_id", w, r)
		project := context.Get(r, "project").(db.Project)

		if err != nil {
			helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid webhook ID",
			})
			return
		}

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

func GetWebhookRefs(w http.ResponseWriter, r *http.Request) {
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)

	if err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid Webhook ID",
		})
		return
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
	log.Info(fmt.Sprintf("Found Project: %v", project.ID))

	if !helpers.Bind(w, r, &webhook) {
		log.Info("Failed to bind for webhook uploads")
		return
	}

	if webhook.ProjectID != project.ID {
		log.Error(fmt.Sprintf("Project ID in body and URL must be the same: %v vs. %v", webhook.ProjectID, project.ID))

		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}
	err := webhook.Validate()
	if err != nil {
		log.Error(err)
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	_, errWebhook := helpers.Store(r).CreateWebhook(webhook)

	if errWebhook != nil {
		log.Error(errWebhook)
		helpers.WriteError(w, errWebhook)
		return
	}

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

	w.WriteHeader(http.StatusNoContent)
}

func DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	webhook_id, err := helpers.GetIntParam("webhook_id", w, r)
	project := context.Get(r, "project").(db.Project)

	err = helpers.Store(r).DeleteWebhook(project.ID, webhook_id)
	if err == db.ErrInvalidOperation {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Webhook failed to be deleted",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
