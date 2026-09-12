package galaxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInstallArgs(t *testing.T) {
	tests := []struct {
		name        string
		installType InstallType
		args        []string
		wantErr     string
	}{
		{"empty", InstallRole, nil, ""},
		{"common flags", InstallRole, []string{"--ignore-errors", "-n", "-vvv"}, ""},
		{"value as separate token", InstallCollection, []string{"--timeout", "60"}, ""},
		{"value inline", InstallCollection, []string{"--server=https://galaxy.example.com/api/?x=1"}, ""},
		{"collection only flags", InstallCollection, []string{"--pre", "-U", "--keyring", "/etc/keyring.gpg"}, ""},
		{"role only flag", InstallRole, []string{"--keep-scm-meta"}, ""},

		{"collection flag on role", InstallRole, []string{"--pre"}, `"--pre" is not allowed`},
		{"role flag on collection", InstallCollection, []string{"-g"}, `"-g" is not allowed`},
		{"token in argv", InstallCollection, []string{"--token", "x"}, `"--token" is not allowed`},
		{"path override", InstallRole, []string{"-p", "/etc"}, `"-p" is not allowed`},
		{"requirements override", InstallRole, []string{"-r", "other.yml"}, `"-r" is not allowed`},
		{"positional argument", InstallRole, []string{"evil.role"}, `"evil.role" is not allowed`},
		{"empty token", InstallRole, []string{""}, `"" is not allowed`},
		{"value for boolean flag", InstallCollection, []string{"--pre=1"}, "does not take a value"},
		{"missing value at end", InstallCollection, []string{"--timeout"}, "requires a value"},
		{"missing value before flag", InstallCollection, []string{"--timeout", "--pre"}, "requires a value"},
		{"empty inline value", InstallCollection, []string{"--timeout="}, "requires a value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstallArgs(tt.installType, tt.args)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
