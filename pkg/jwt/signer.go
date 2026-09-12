package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	josejwt "github.com/go-jose/go-jose/v4/jwt"
)

// Signer mints signed JWTs and exposes its public key as a JWKS endpoint.
type Signer interface {
	// Sign produces a serialized compact JWS for the given task.
	Sign(info TaskInfo) (string, error)
	// JWKS returns the JSON Web Key Set containing the current public key.
	JWKS() ([]byte, error)
	// KeyID returns the kid for the current signing key.
	KeyID() string
}

type ecdsaSigner struct {
	mu      sync.Mutex
	key     *ecdsa.PrivateKey
	kid     string
	signer  jose.Signer
	options SignerOptions
}

// GenerateKeyPEM generates a new ECDSA P-256 private key encoded as PKCS#8 PEM.
func GenerateKeyPEM() ([]byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("jwt: generate key: %w", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("jwt: marshal key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}), nil
}

const defaultTTL = 24 * time.Hour

// NewECDSASignerFromPEM creates a Signer from a PEM-encoded ECDSA P-256 private key.
func NewECDSASignerFromPEM(pemBytes []byte, opts SignerOptions) (Signer, error) {
	if opts.MaxTTL <= 0 {
		opts.MaxTTL = defaultTTL
	}
	if opts.DefaultTTL <= 0 {
		opts.DefaultTTL = defaultTTL
	}
	if opts.DefaultTTL > opts.MaxTTL {
		opts.DefaultTTL = opts.MaxTTL
	}

	key, err := parsePrivateKey(pemBytes)
	if err != nil {
		return nil, err
	}

	kid, err := computeKID(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	signingKey := jose.SigningKey{
		Algorithm: jose.ES256,
		Key: jose.JSONWebKey{
			Key:       key,
			KeyID:     kid,
			Algorithm: string(jose.ES256),
			Use:       "sig",
		},
	}
	sigOpts := (&jose.SignerOptions{}).WithType("JWT")
	js, err := jose.NewSigner(signingKey, sigOpts)
	if err != nil {
		return nil, fmt.Errorf("jwt: create signer: %w", err)
	}

	return &ecdsaSigner{
		key:     key,
		kid:     kid,
		signer:  js,
		options: opts,
	}, nil
}

func (s *ecdsaSigner) KeyID() string { return s.kid }

func (s *ecdsaSigner) Sign(info TaskInfo) (string, error) {
	now := time.Now().UTC()

	jti, err := randomJTI()
	if err != nil {
		return "", err
	}

	ttl := s.options.DefaultTTL
	if info.TTL > 0 {
		ttl = info.TTL
	}
	if s.options.MaxTTL > 0 && ttl > s.options.MaxTTL {
		ttl = s.options.MaxTTL
	}

	claims := TaskClaims{
		Issuer:    s.options.Issuer,
		Audience:  info.Audience,
		Subject:   fmt.Sprintf("task:%d", info.TaskID),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
		JWTID:     jti,

		TaskID:     info.TaskID,
		ProjectID:  info.ProjectID,
		TemplateID: info.TemplateID,
		UserID:     info.UserID,
	}

	return josejwt.Signed(s.signer).Claims(claims).Serialize()
}

func (s *ecdsaSigner) JWKS() ([]byte, error) {
	jwk := jose.JSONWebKey{
		Key:       &s.key.PublicKey,
		KeyID:     s.kid,
		Algorithm: string(jose.ES256),
		Use:       "sig",
	}
	set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
	return json.Marshal(set)
}

func parsePrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("jwt: no PEM block found in key file")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("jwt: key is not ECDSA (got %T)", key)
	}
	if ecKey.Curve != elliptic.P256() {
		return nil, fmt.Errorf("jwt: unsupported curve %q, expected P-256", ecKey.Curve.Params().Name)
	}
	return ecKey, nil
}

func computeKID(pub *ecdsa.PublicKey) (string, error) {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(der)
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

func randomJTI() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}
