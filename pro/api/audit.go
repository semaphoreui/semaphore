package api

import (
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

const auditProOnly = "Filters and export of the audit log are available in Semaphore PRO"

type AuditController struct{}

func NewAuditController(db.Store, pro_interfaces.SubscriptionService) *AuditController {
	return &AuditController{}
}

// GetEvents answers a filtered page of the audit log, filters are Pro.
func (c *AuditController) GetEvents(w http.ResponseWriter, r *http.Request) {
	helpers.WriteErrorStatus(w, auditProOnly, http.StatusForbidden)
}

func (c *AuditController) ExportEvents(w http.ResponseWriter, r *http.Request) {
	helpers.WriteErrorStatus(w, auditProOnly, http.StatusForbidden)
}
