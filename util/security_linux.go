//go:build linux

package util

import (
	"golang.org/x/sys/unix"
)

// SetNoNewPrivs sets the no_new_privs flag on the current process.
// Once set, this is inherited by all child processes and cannot be unset.
// It prevents privilege escalation via setuid/setgid binaries and
// filesystem capabilities.
func SetNoNewPrivs() error {
	return unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0)
}
