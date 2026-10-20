//go:build integration

package integration

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// certFiles is a PEM certificate and its key on disk, the form Mailpit wants.
type certFiles struct {
	certPath, keyPath string
}

// testCerts is a private CA with one server certificate signed by it and an
// unrelated self-signed certificate for the "untrusted issuer" scenario. Both
// server certificates carry the same names, so only the issuer differs.
type testCerts struct {
	pool      *x509.CertPool
	server    certFiles
	untrusted certFiles
}

// serverNames are the SANs of every test server certificate: the loopback
// names Testcontainers exposes the port on, plus a fake remote host used to
// exercise the "not localhost" branches without DNS.
var serverNames = []string{"localhost", "smtp.test"}

func newTestCerts(t *testing.T) *testCerts {
	t.Helper()
	dir := t.TempDir()

	caKey, caCert := newCA(t)
	pool := x509.NewCertPool()
	pool.AddCert(caCert)

	return &testCerts{
		pool:      pool,
		server:    writeServerCert(t, dir, "server", caKey, caCert),
		untrusted: writeServerCert(t, dir, "untrusted", nil, nil),
	}
}

func newCA(t *testing.T) (*ecdsa.PrivateKey, *x509.Certificate) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Semaphore test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return key, cert
}

// writeServerCert issues a certificate for serverNames. With a nil issuer
// the certificate is self-signed.
func writeServerCert(t *testing.T, dir, name string, caKey *ecdsa.PrivateKey, caCert *x509.Certificate) certFiles {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "smtp.test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     serverNames,
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	parent, signer := tmpl, crypto.Signer(key)
	if caCert != nil {
		parent, signer = caCert, caKey
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, parent, &key.PublicKey, signer)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	files := certFiles{
		certPath: filepath.Join(dir, name+".crt"),
		keyPath:  filepath.Join(dir, name+".key"),
	}
	require.NoError(t, os.WriteFile(files.certPath,
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644))
	require.NoError(t, os.WriteFile(files.keyPath,
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o644))
	return files
}
