package helpers

import (
	"net/http"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	log "github.com/sirupsen/logrus"
)

func RecordResourceEvent(r *http.Request, event audit.ResourceEvent) {
	service, ok := GetFromContext(r, "audit_service").(*audit.Service)
	if !ok || service == nil {
		log.Error("Failed to record resource event: missing audit service")
		return
	}
	requestContext, ok := AuditRequestContextFrom(r)
	if !ok {
		log.Error("Failed to record resource event: missing audit request context")
		return
	}
	user, ok := GetFromContext(r, "user").(*db.User)
	if !ok || user == nil {
		log.Error("Failed to record resource event: missing authenticated user")
		return
	}

	event.Actor = audit.Actor{ID: user.ID, Name: user.Username}
	event.Request = audit.Request{
		ID:        requestContext.RequestID,
		SourceIP:  requestContext.SourceIP,
		UserAgent: requestContext.UserAgent,
	}
	if err := service.RecordResource(event); err != nil {
		log.WithError(err).Error("Failed to record resource event")
	}
}
