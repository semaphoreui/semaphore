package util

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *AuditConfig
		wantErr string
	}{
		{"nil section", nil, ""},
		{"disabled section is ignored", &AuditConfig{TrustedProxyCIDRs: []string{"junk"}}, ""},
		{"enabled without instance id", &AuditConfig{Enabled: true}, "audit.instance_id"},
		{"instance id with a space", &AuditConfig{Enabled: true, InstanceID: "prod eu"}, "audit.instance_id"},
		{"instance id too long", &AuditConfig{Enabled: true, InstanceID: strings.Repeat("a", 256)}, "audit.instance_id"},
		{"instance id not ascii", &AuditConfig{Enabled: true, InstanceID: "прод"}, "audit.instance_id"},
		{"invalid cidr", &AuditConfig{Enabled: true, InstanceID: "prod-eu", TrustedProxyCIDRs: []string{"10.0.0.0/33"}}, "trusted_proxy_cidrs"},
		{"ipv4-mapped network shorter than /96", &AuditConfig{Enabled: true, InstanceID: "prod-eu", TrustedProxyCIDRs: []string{"::ffff:0:0/80"}}, "/96"},
		{"negative retention", &AuditConfig{Enabled: true, InstanceID: "prod-eu", RetentionDays: -1}, "audit.retention_days"},
		{"retention set", &AuditConfig{Enabled: true, InstanceID: "prod-eu", RetentionDays: 90}, ""},
		{"valid", &AuditConfig{Enabled: true, InstanceID: "prod-eu", TrustedProxyCIDRs: []string{"10.0.0.0/8", "fd00::/8"}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestAuditConfig_TrustedProxies(t *testing.T) {
	prefixes, err := (&AuditConfig{TrustedProxyCIDRs: []string{" 10.1.2.3/8", "fd00::1/8"}}).TrustedProxies()
	require.NoError(t, err)
	assert.Equal(t, []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("fd00::/8")}, prefixes)
}

func TestAuditConfig_IsEnabled(t *testing.T) {
	var nilConfig *AuditConfig
	assert.False(t, nilConfig.IsEnabled())
	assert.True(t, (&AuditConfig{Enabled: true}).IsEnabled())
}

func TestAuditSyslogConfig_IsConfigured(t *testing.T) {
	var nilConfig *AuditSyslogConfig
	assert.False(t, nilConfig.IsConfigured())
	assert.False(t, (&AuditSyslogConfig{Timeout: "10s"}).IsConfigured(), "defaults alone do not configure it")
	assert.False(t, (&AuditSyslogConfig{CAFile: "/etc/ssl/ca.pem", ServerName: "siem"}).IsConfigured(), "optional fields alone send nothing")
	assert.True(t, (&AuditSyslogConfig{Address: "siem:6514"}).IsConfigured())
}

func TestAuditConfig_FromEnvironment(t *testing.T) {
	t.Setenv("SEMAPHORE_AUDIT_ENABLED", "true")
	t.Setenv("SEMAPHORE_AUDIT_INSTANCE_ID", "prod-eu")
	t.Setenv("SEMAPHORE_AUDIT_TRUSTED_PROXY_CIDRS", `["10.0.0.0/8","fd00::/8"]`)
	t.Setenv("SEMAPHORE_AUDIT_SYSLOG_ID", "siem-syslog")
	t.Setenv("SEMAPHORE_AUDIT_SYSLOG_ADDRESS", "siem.example:6514")

	config := &ConfigType{}
	_, err := loadEnvironmentToObject(config)
	require.NoError(t, err)
	require.NoError(t, loadDefaultsToObject(config))

	assert.True(t, config.Audit.Enabled)
	assert.Equal(t, "prod-eu", config.Audit.InstanceID)
	assert.Equal(t, []string{"10.0.0.0/8", "fd00::/8"}, config.Audit.TrustedProxyCIDRs)
	assert.Equal(t, "siem-syslog", config.Audit.Syslog.ID)
	assert.Equal(t, "siem.example:6514", config.Audit.Syslog.Address)
	assert.Equal(t, "10s", config.Audit.Syslog.Timeout)
	assert.NoError(t, config.Audit.Validate())
}

func TestAuditSplunkHECConfig_IsConfigured(t *testing.T) {
	var nilConfig *AuditSplunkHECConfig
	assert.False(t, nilConfig.IsConfigured())
	assert.False(t, (&AuditSplunkHECConfig{Source: "semaphore", Sourcetype: "semaphore:audit", Timeout: "10s"}).IsConfigured(), "defaults alone do not configure it")
	assert.False(t, (&AuditSplunkHECConfig{Token: "t", Index: "security"}).IsConfigured(), "optional fields alone send nothing")
	assert.True(t, (&AuditSplunkHECConfig{URL: "https://hec.example:8088/services/collector/event"}).IsConfigured())
	assert.True(t, (&AuditSplunkHECConfig{ID: "siem-hec"}).IsConfigured())
}

func TestAuditSplunkHECConfig_FromEnvironment(t *testing.T) {
	t.Setenv("SEMAPHORE_AUDIT_ENABLED", "true")
	t.Setenv("SEMAPHORE_AUDIT_INSTANCE_ID", "prod-eu")
	t.Setenv("SEMAPHORE_AUDIT_SPLUNK_HEC_ID", "siem-hec")
	t.Setenv("SEMAPHORE_AUDIT_SPLUNK_HEC_URL", "https://hec.example:8088/services/collector/event")
	t.Setenv("SEMAPHORE_AUDIT_SPLUNK_HEC_TOKEN", "00000000-0000-0000-0000-000000000000")
	t.Setenv("SEMAPHORE_AUDIT_SPLUNK_HEC_INDEX", "security")

	config := &ConfigType{}
	sensitive, err := loadEnvironmentToObject(config)
	require.NoError(t, err)
	assert.Contains(t, sensitive, "SEMAPHORE_AUDIT_SPLUNK_HEC_TOKEN")
	require.NoError(t, loadDefaultsToObject(config))

	hec := config.Audit.SplunkHEC
	require.NotNil(t, hec)
	assert.Equal(t, "siem-hec", hec.ID)
	assert.Equal(t, "https://hec.example:8088/services/collector/event", hec.URL)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", hec.Token)
	assert.Equal(t, "security", hec.Index)
	assert.Equal(t, "semaphore", hec.Source)
	assert.Equal(t, "semaphore:audit", hec.Sourcetype)
	assert.Equal(t, "10s", hec.Timeout)
}

func TestAuditConfig_TrustedProxiesUnmapsIPv4MappedNetworks(t *testing.T) {
	prefixes, err := (&AuditConfig{TrustedProxyCIDRs: []string{
		"::ffff:10.0.0.0/104", "::ffff:192.168.1.7/128", "10.1.0.0/16",
	}}).TrustedProxies()

	require.NoError(t, err)
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("192.168.1.7/32"),
		netip.MustParsePrefix("10.1.0.0/16"),
	}, prefixes)
	assert.True(t, prefixes[0].Contains(netip.MustParseAddr("10.1.2.3")), "an unmapped peer matches")
}

func TestAuditConfig_RetentionFromEnvironment(t *testing.T) {
	t.Setenv("SEMAPHORE_AUDIT_RETENTION_DAYS", "180")
	config := &ConfigType{}
	_, err := loadEnvironmentToObject(config)
	require.NoError(t, err)
	require.NotNil(t, config.Audit)
	assert.Equal(t, 180, config.Audit.RetentionDays)
}
