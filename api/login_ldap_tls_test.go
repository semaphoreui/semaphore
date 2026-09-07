package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTestCA generates a throwaway self-signed CA and returns its PEM path.
func writeTestCA(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE", Bytes: der,
	}), 0600))

	return path
}

func TestLdapTLSConfig_DefaultsToVerify(t *testing.T) {
	cfg, err := ldapTLSConfig(util.LdapProvider{})

	require.NoError(t, err)
	assert.False(t, cfg.InsecureSkipVerify, "a configured provider must verify by default")
	assert.Nil(t, cfg.RootCAs, "no CA file means the system trust store is used")
}

// The legacy flat ldap_* config is synthesized with TLSSkipVerify set, so that
// upgrading installs on self-signed certs keep working until they opt in.
func TestLdapTLSConfig_SkipVerifyOptOut(t *testing.T) {
	cfg, err := ldapTLSConfig(util.LdapProvider{TLSSkipVerify: true})

	require.NoError(t, err)
	assert.True(t, cfg.InsecureSkipVerify)
	assert.Nil(t, cfg.RootCAs)
}

func TestLdapTLSConfig_CACertFileOverridesSkipVerify(t *testing.T) {
	cfg, err := ldapTLSConfig(util.LdapProvider{TLSSkipVerify: true, TLSCACertFile: writeTestCA(t)})

	require.NoError(t, err)
	assert.False(t, cfg.InsecureSkipVerify, "supplying a CA forces verification on")
	assert.NotNil(t, cfg.RootCAs)
}

func TestLdapTLSConfig_CACertFileEnablesVerification(t *testing.T) {
	cfg, err := ldapTLSConfig(util.LdapProvider{TLSCACertFile: writeTestCA(t)})

	require.NoError(t, err)
	assert.False(t, cfg.InsecureSkipVerify, "supplying a CA implies opting into verification")
	assert.NotNil(t, cfg.RootCAs)
}

func TestLdapTLSConfig_MissingCACertFile(t *testing.T) {
	_, err := ldapTLSConfig(util.LdapProvider{TLSCACertFile: "/nonexistent/ca.pem"})

	require.Error(t, err)
	assert.ErrorContains(t, err, "read ldap CA cert file")
}

func TestLdapTLSConfig_InvalidPEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "garbage.pem")
	require.NoError(t, os.WriteFile(path, []byte("not a certificate"), 0600))

	_, err := ldapTLSConfig(util.LdapProvider{TLSCACertFile: path})

	require.Error(t, err)
	assert.ErrorContains(t, err, "no valid PEM certificates")
}
