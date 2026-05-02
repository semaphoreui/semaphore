//go:build !linux

package util

// SetNoNewPrivs is a no-op on non-Linux platforms.
func SetNoNewPrivs() error {
	return nil
}
