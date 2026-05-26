package jwt

import "sync"

var (
	defaultMu     sync.RWMutex
	defaultSigner Signer
)

// SetDefault registers the process-wide Signer. Callers (typically the server
// bootstrap) call this once after building the signer from config. Passing
// nil clears the default (e.g. when JWT issuance is disabled).
func SetDefault(s Signer) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultSigner = s
}

// Default returns the registered Signer or nil when JWT issuance is disabled.
func Default() Signer {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultSigner
}
