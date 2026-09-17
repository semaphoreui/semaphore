package pro_interfaces

import "context"

// AuditExporter owns asynchronous audit export processing.
type AuditExporter interface {
	// Start is non-blocking: it starts internal processing and returns after startup.
	// It may return an error after partial initialization; the owner may still call Stop.
	Start(context.Context) error
	// Stop initiates cancellation, interrupts context-aware network operations, and waits
	// for every exporter goroutine to finish. It must not return while any goroutine can
	// access the audit store, even when it returns a cleanup error. Stop is safe after a
	// partial Start and on repeated calls.
	Stop(context.Context) error
}
