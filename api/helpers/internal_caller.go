package helpers

import (
	"context"

	"github.com/semaphoreui/semaphore/services/audit"
)

// InternalCaller is the user on whose behalf a server component calls the REST
// API in-process. The component signed the user in with its own credential, so
// the request carries none; authentication takes the user from here.
// Only server code can set it: context values never come from the network.
type InternalCaller struct {
	UserID int
	// Auth is the component's sign-in method, the actor's auth method in audit.
	Auth audit.AuthMethod
	// TokenFingerprint is audit.TokenFingerprint of the component's credential.
	TokenFingerprint string
}

type internalCallerKey struct{}

// WithInternalCaller marks ctx as an in-process call on behalf of caller.
func WithInternalCaller(ctx context.Context, caller InternalCaller) context.Context {
	return context.WithValue(ctx, internalCallerKey{}, caller)
}

// InternalCallerFrom returns the caller of an in-process call.
func InternalCallerFrom(ctx context.Context) (InternalCaller, bool) {
	caller, ok := ctx.Value(internalCallerKey{}).(InternalCaller)
	return caller, ok
}
