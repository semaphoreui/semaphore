package db

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type HostConfigType string

const (
	// HostConfigHost matches an SSH host name and is applied as an ssh_config
	// Host block.
	HostConfigHost HostConfigType = "host"

	// HostConfigURL matches a git URL or a URL prefix and is applied as a git
	// url.<replacement>.insteadOf rewrite.
	HostConfigURL HostConfigType = "url"
)

// HostConfig maps an SSH host or a git URL to the credential git must use to
// reach it. Without a mapping only the key of the repository itself is
// available, which leaves submodules, galaxy requirements and inventory
// repositories on other hosts without credentials.
type HostConfig struct {
	ID        int            `db:"id" json:"id" backup:"-"`
	ProjectID int            `db:"project_id" json:"project_id" backup:"-"`
	Type      HostConfigType `db:"type" json:"type"`

	// Name is the ssh host name for a host mapping, or the URL/prefix for a URL
	// mapping: it is what identifies a mapping. It is called Name because the
	// shared reference lookup reads that field and column, and stays "host" in
	// the API, which is the word the interface uses.
	Name string `db:"name" json:"host" binding:"required"`

	SSHKeyID int       `db:"ssh_key_id" json:"ssh_key_id" binding:"required" backup:"-"`
	SSHKey   AccessKey `db:"-" json:"-" backup:"-"`
}

// hostConfigHostRE matches a DNS name, optionally with a port. An allowlist,
// because the host is written into a generated ssh_config and into a git
// url.<x>.insteadOf key: anything outside this set is config or shell syntax.
var hostConfigHostRE = regexp.MustCompile(
	`^[A-Za-z0-9_]([A-Za-z0-9_-]*[A-Za-z0-9_])?(\.[A-Za-z0-9_]([A-Za-z0-9_-]*[A-Za-z0-9_])?)*$`)

// ValidateMapping checks what a mapping points at: its type, and the host or URL
// which ends up in a generated ssh config and a git rewrite.
//
// Separate from Validate because a backup carries the credential by name, so the
// key id is only known once the backup is restored. The preflight pass still has
// to reject a malformed host before anything is written.
func (h *HostConfig) ValidateMapping() error {
	h.Name = strings.TrimSpace(h.Name)

	if h.Name == "" {
		return common_errors.NewValidationError("host or URL can not be empty")
	}

	switch h.Type {
	case HostConfigHost:
		return h.validateHost()
	case HostConfigURL:
		return h.validateURL()
	default:
		return common_errors.NewValidationError("unsupported mapping type")
	}
}

func (h *HostConfig) Validate() error {
	if err := h.ValidateMapping(); err != nil {
		return err
	}

	if h.SSHKeyID <= 0 {
		return common_errors.NewValidationError("a credential must be selected")
	}

	return nil
}

func (h *HostConfig) validateHost() error {
	if !hostConfigHostRE.MatchString(h.Name) {
		return common_errors.NewValidationError("host must be a host name, for example github.com")
	}
	return nil
}

func (h *HostConfig) validateURL() error {
	u, err := url.Parse(h.Name)
	if err != nil || u.Host == "" {
		return common_errors.NewValidationError("URL must be a git URL, for example https://github.com/acme/")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return common_errors.NewValidationError("URL must start with http:// or https://")
	}

	if u.User != nil {
		return common_errors.NewValidationError("URL must not contain credentials")
	}

	// The URL ends up in a GIT_CONFIG_PARAMETERS entry, which git splits at the
	// first "=", and in a config key, which ends at whitespace. An "=" can not be
	// escaped out of the problem: percent-encoding it stops the rewrite matching
	// the real URL, so the mapping would silently do nothing.
	if strings.ContainsAny(h.Name, " \t\r\n\"'\\$`=") {
		return common_errors.NewValidationError("URL contains invalid characters")
	}

	if !hostConfigHostRE.MatchString(u.Hostname()) {
		return common_errors.NewValidationError("URL contains an invalid host name")
	}

	return nil
}

// SSHAlias is the ssh_config Host name a URL mapping rewrites to. A URL mapping
// can not use the real host name: several mappings may share one host and each
// needs its own credential, so every mapping gets its own alias.
// ValidateCredential checks the mapping against the kind of credential it
// points at.
//
// A host mapping becomes an ssh_config Host block, which can only bind an ssh
// identity. A URL mapping rewrites the URL, so it can also carry a login and a
// password — but only over https, because the rewrite puts them in the URL and
// nothing else keeps them off the wire.
func (h *HostConfig) ValidateCredential(keyType AccessKeyType) error {
	switch h.Type {
	case HostConfigHost:
		if keyType != AccessKeySSH {
			return common_errors.NewValidationError("a host mapping needs an SSH key")
		}
	case HostConfigURL:
		switch keyType {
		case AccessKeySSH:
		case AccessKeyLoginPassword:
			if !strings.HasPrefix(strings.TrimSpace(h.Name), "https://") {
				return common_errors.NewValidationError(
					"a login/password credential can only be used with an https URL")
			}
		default:
			return common_errors.NewValidationError(
				"a URL mapping needs an SSH key or a login/password credential")
		}
	}

	return nil
}

func (h HostConfig) SSHAlias() string {
	return "semaphore-mapping-" + strconv.Itoa(h.ID)
}
