package util

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"strings"
)

type AuditConfig struct {
	// Enabled turns on the security audit log.
	Enabled bool `json:"enabled,omitempty" env:"SEMAPHORE_AUDIT_ENABLED"`
	// InstanceID identifies this installation in audit events. Required when enabled.
	InstanceID string `json:"instance_id,omitempty" env:"SEMAPHORE_AUDIT_INSTANCE_ID"`
	// TrustedProxyCIDRs lists proxies whose X-Forwarded-For and X-Real-IP are trusted.
	TrustedProxyCIDRs []string `json:"trusted_proxy_cidrs,omitempty" env:"SEMAPHORE_AUDIT_TRUSTED_PROXY_CIDRS"`
	// Syslog exports audit events over Syslog+TLS.
	Syslog *AuditSyslogConfig `json:"syslog,omitempty"`
}

type AuditSyslogConfig struct {
	// ID keys the export cursor. A new ID starts from now.
	ID string `json:"id,omitempty" env:"SEMAPHORE_AUDIT_SYSLOG_ID"`
	// Address is the receiver host:port.
	Address string `json:"address,omitempty" env:"SEMAPHORE_AUDIT_SYSLOG_ADDRESS"`
	// Timeout bounds connecting and writing one batch.
	Timeout string `json:"timeout,omitempty" default:"10s" env:"SEMAPHORE_AUDIT_SYSLOG_TIMEOUT"`
	// CAFile is a PEM bundle added to the system roots.
	CAFile string `json:"ca_file,omitempty" env:"SEMAPHORE_AUDIT_SYSLOG_CA_FILE"`
	// ServerName overrides the name checked in the receiver certificate.
	ServerName string `json:"server_name,omitempty" env:"SEMAPHORE_AUDIT_SYSLOG_SERVER_NAME"`
}

func (c *AuditConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

func (c *AuditConfig) Validate() error {
	if !c.IsEnabled() {
		return nil
	}
	if !regexp.MustCompile(`^[!-~]{1,255}$`).MatchString(c.InstanceID) {
		return errors.New("audit.instance_id is required when audit is enabled: 1-255 printable ASCII characters without spaces")
	}
	_, err := c.TrustedProxies()
	return err
}

func (c *AuditConfig) TrustedProxies() ([]netip.Prefix, error) {
	if c == nil {
		return nil, nil
	}
	prefixes := make([]netip.Prefix, 0, len(c.TrustedProxyCIDRs))
	for _, cidr := range c.TrustedProxyCIDRs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(cidr))
		if err != nil {
			return nil, fmt.Errorf("audit.trusted_proxy_cidrs: invalid CIDR %q: %w", cidr, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func (c *AuditSyslogConfig) IsConfigured() bool {
	return c != nil && (c.ID != "" || c.Address != "")
}
