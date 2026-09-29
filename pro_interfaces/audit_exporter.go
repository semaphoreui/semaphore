package pro_interfaces

type AuditExporter interface {
	// Start creates destination cursors. An error stops the server.
	Start() error
	Stop()
	DestinationIDs() []string
}
