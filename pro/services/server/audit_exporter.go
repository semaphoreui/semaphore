package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

type noopAuditExporter struct{}

var _ pro_interfaces.AuditExporter = (*noopAuditExporter)(nil)

func NewAuditExporter(_ db.Store) (pro_interfaces.AuditExporter, error) {
	return &noopAuditExporter{}, nil
}

func (*noopAuditExporter) Stop() {}
