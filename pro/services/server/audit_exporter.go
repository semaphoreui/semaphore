package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/metrics"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
)

func NewAuditExporter(_ db.Store, _ *util.AuditConfig, _ pro_interfaces.AuditExportLeaser, _ *metrics.Metrics) pro_interfaces.AuditExporter {
	return auditExporterStub{}
}

type auditExporterStub struct{}

func (auditExporterStub) Start() error             { return nil }
func (auditExporterStub) Stop()                    {}
func (auditExporterStub) DestinationIDs() []string { return []string{} }
