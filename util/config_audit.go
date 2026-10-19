package util

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"strings"
)

type AuditConfig struct {
	// Enabled turns on the audit log.
	Enabled bool `json:"enabled,omitempty" env:"SEMAPHORE_AUDIT_ENABLED"`
	// InstanceID is the installation name added to every event. Required when enabled.
	InstanceID string `json:"instance_id,omitempty" env:"SEMAPHORE_AUDIT_INSTANCE_ID"`
	// TrustedProxyCIDRs lists the proxy networks allowed to pass the client address.
	TrustedProxyCIDRs []string `json:"trusted_proxy_cidrs,omitempty" env:"SEMAPHORE_AUDIT_TRUSTED_PROXY_CIDRS"`
	// RetentionDays deletes audit events older than this many days. 0 keeps everything.
	RetentionDays int `json:"retention_days,omitempty" env:"SEMAPHORE_AUDIT_RETENTION_DAYS"`
	// Syslog sends audit events to a Syslog receiver over TLS.
	Syslog *AuditSyslogConfig `json:"syslog,omitempty"`
	// SplunkHEC exports audit events over the Splunk HTTP Event Collector protocol.
	SplunkHEC *AuditSplunkHECConfig `json:"splunk_hec,omitempty"`
}

type AuditSyslogConfig struct {
	// ID names the destination. Keep it when the address changes.
	ID string `json:"id,omitempty" env:"SEMAPHORE_AUDIT_SYSLOG_ID"`
	// Address is the receiver host and port.
	Address string `json:"address,omitempty" env:"SEMAPHORE_AUDIT_SYSLOG_ADDRESS"`
	// Timeout limits connecting to the receiver and sending events.
	Timeout string `json:"timeout,omitempty" default:"10s" env:"SEMAPHORE_AUDIT_SYSLOG_TIMEOUT"`
	// CAFile is a PEM file with extra CA certificates to trust.
	CAFile string `json:"ca_file,omitempty" env:"SEMAPHORE_AUDIT_SYSLOG_CA_FILE"`
	// ServerName is the name to check in the receiver certificate.
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
	if c.RetentionDays < 0 {
		return errors.New("audit.retention_days must be 0 to keep every event, or a number of days")
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
		if prefix.Addr().Is4In6() && prefix.Bits() < 96 {
			return nil, fmt.Errorf("audit.trusted_proxy_cidrs: IPv4-mapped network %q needs a prefix of /96 or longer", cidr)
		}
		prefix = prefix.Masked()
		// Peers are compared unmapped, so a mapped IPv4 network must be unmapped too.
		if prefix.Addr().Is4In6() && prefix.Bits() >= 96 {
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func (c *AuditSyslogConfig) IsConfigured() bool {
	return c != nil && (c.ID != "" || c.Address != "")
}

type AuditSplunkHECConfig struct {
	// ID names the destination. Keep it when the URL changes. A new ID starts with new events.
	ID string `json:"id,omitempty" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_ID"`
	// URL is the full HEC endpoint, for example https://hec.example:8088/services/collector/event.
	URL string `json:"url,omitempty" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_URL"`
	// Token is the HEC token sent as "Authorization: Splunk <token>".
	Token string `json:"token,omitempty" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_TOKEN,sensitive"`
	// Index is the target index. Empty uses the default index of the token.
	Index string `json:"index,omitempty" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_INDEX"`
	// Source is the HEC source of every event.
	Source string `json:"source,omitempty" default:"semaphore" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_SOURCE"`
	// Sourcetype is the HEC sourcetype of every event.
	Sourcetype string `json:"sourcetype,omitempty" default:"semaphore:audit" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_SOURCETYPE"`
	// Timeout bounds one HTTP request with one batch of events.
	Timeout string `json:"timeout,omitempty" default:"10s" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_TIMEOUT"`
	// CAFile is a PEM bundle added to the system roots.
	CAFile string `json:"ca_file,omitempty" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_CA_FILE"`
	// ServerName overrides the name checked in the receiver certificate.
	ServerName string `json:"server_name,omitempty" env:"SEMAPHORE_AUDIT_SPLUNK_HEC_SERVER_NAME"`
}

func (c *AuditSplunkHECConfig) IsConfigured() bool {
	return c != nil && (c.ID != "" || c.URL != "")
}
