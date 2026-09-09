package galaxy

import (
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// InstallType selects the `ansible-galaxy <type> install` subcommand.
type InstallType string

const (
	InstallRole       InstallType = "role"
	InstallCollection InstallType = "collection"
)

// installFlags maps each allowed `ansible-galaxy <type> install` flag to
// whether it takes a value. The two subcommands accept different flags, so the
// set depends on installType.
//
// Deliberately excluded: --token/--api-key (secret in argv, visible in the
// process list), -p/--roles-path/--collections-path (write outside the
// repository), -r (set by Semaphore).
func installFlags(installType InstallType) map[string]bool {
	flags := map[string]bool{
		"-c":                false,
		"--ignore-certs":    false,
		"-f":                false,
		"--force":           false,
		"--force-with-deps": false,
		"-i":                false,
		"--ignore-errors":   false,
		"-n":                false,
		"--no-deps":         false,
		"-s":                true,
		"--server":          true,
		"--timeout":         true,
		"-v":                false,
		"-vv":               false,
		"-vvv":              false,
		"-vvvv":             false,
		"--verbose":         false,
	}

	switch installType {
	case InstallRole:
		flags["-g"] = false
		flags["--keep-scm-meta"] = false
	case InstallCollection:
		flags["--pre"] = false
		flags["-U"] = false
		flags["--upgrade"] = false
		flags["--offline"] = false
		flags["--no-cache"] = false
		flags["--clear-response-cache"] = false
		flags["--disable-gpg-verify"] = false
		flags["--keyring"] = true
		flags["--signature"] = true
		flags["--required-valid-signature-count"] = true
		flags["--ignore-signature-status-code"] = true
		flags["--ignore-signature-status-codes"] = true
	}

	return flags
}

// ValidateInstallArgs checks that every argv token in args is an allowed flag
// for `ansible-galaxy <installType> install`. Both `--flag=value` and
// `--flag value` forms are accepted for flags that take a value.
func ValidateInstallArgs(installType InstallType, args []string) error {
	allowed := installFlags(installType)
	prefix := "ansible-galaxy " + string(installType) + " install: "

	for i := 0; i < len(args); i++ {
		flag, value, hasInlineValue := strings.Cut(args[i], "=")

		takesValue, ok := allowed[flag]
		if !ok {
			return common_errors.NewValidationError(prefix + "argument \"" + args[i] + "\" is not allowed")
		}

		if !takesValue {
			if hasInlineValue {
				return common_errors.NewValidationError(prefix + "flag " + flag + " does not take a value")
			}
			continue
		}

		if !hasInlineValue {
			i++
			if i < len(args) {
				value = args[i]
			}
		}

		if value == "" || strings.HasPrefix(value, "-") {
			return common_errors.NewValidationError(prefix + "flag " + flag + " requires a value")
		}
	}

	return nil
}
