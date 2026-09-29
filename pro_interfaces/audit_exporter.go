package pro_interfaces

import "context"

type AuditExporter interface {
	// Start creates destination cursors. An error stops the server.
	Start() error
	Stop()
	DestinationIDs() []string
}

// AuditExportLeaser lets one node at a time export a destination.
type AuditExportLeaser interface {
	TryAcquire(ctx context.Context, destinationID string) (lease AuditExportLease, ok bool, err error)
}

type AuditExportLease interface {
	// Lost fires when the lease can no longer be trusted.
	Lost() <-chan struct{}
	Release()
}
