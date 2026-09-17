package server

import (
	"errors"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
)

// NewAuditExporter constructs an exporter without starting workers. On error it leaves no
// goroutines or long-lived resources behind.
func NewAuditExporter(
	_ util.AuditConfig,
	_ db.AuditExporterStore,
) (pro_interfaces.AuditExporter, error) {
	return nil, errors.New("audit export is only available in the proprietary build")
}
