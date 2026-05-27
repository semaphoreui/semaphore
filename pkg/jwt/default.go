package jwt

import "sync"

var (
	defaultMu     sync.RWMutex
	defaultSigner Signer
)

// SetDefault registers the global Signer.
func SetDefault(s Signer) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultSigner = s
}

// Default returns the global Signer or nil when JWT issuance is disabled.
func Default() Signer {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultSigner
}
