package pro_interfaces

// AuditExporter owns background delivery of persisted audit events.
type AuditExporter interface {
	Stop()
}
