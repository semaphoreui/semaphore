package jwt

import (
	"crypto/rand"
	"crypto/rsa"
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

// Signer mints signed JWTs and exposes its public key as a JWKS document.
type Signer interface {
	// Sign produces a serialized compact JWS for the given task. It applies
	// the configured Issuer, Audience and TTL.
	Sign(info TaskInfo) (string, error)
	// JWKS returns the JSON Web Key Set containing the current public key.
	JWKS() ([]byte, error)
	// KeyID returns the kid for the current signing key.
	KeyID() string
}

type rsaSigner struct {
	mu      sync.Mutex
	key     *rsa.PrivateKey
	kid     string
	signer  jose.Signer
	options SignerOptions
}

// GenerateKeyPEM generates a new 2048-bit RSA private key and returns it as a
// PKCS#1 PEM-encoded byte slice. Use this when bootstrapping a new signing key
// that will be stored (encrypted) in the database.
func GenerateKeyPEM() ([]byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("jwt: generate key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), nil
}

// NewRSASignerFromPEM creates a Signer from a PEM-encoded RSA private key that
// the caller has already loaded (and decrypted, if stored encrypted). This is
// the primary constructor used by the server; key management (generation,
// encryption, storage) is handled by the util layer.
func NewRSASignerFromPEM(pemBytes []byte, opts SignerOptions) (Signer, error) {
	if opts.TTL <= 0 {
		opts.TTL = time.Hour
	}
	const maxTTL = 24 * time.Hour
	if opts.TTL > maxTTL {
		opts.TTL = maxTTL
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
		Algorithm: jose.RS256,
		Key:       jose.JSONWebKey{Key: key, KeyID: kid, Algorithm: string(jose.RS256), Use: "sig"},
	}
	sigOpts := (&jose.SignerOptions{}).WithType("JWT")
	js, err := jose.NewSigner(signingKey, sigOpts)
	if err != nil {
		return nil, fmt.Errorf("jwt: create signer: %w", err)
	}

	return &rsaSigner{
		key:     key,
		kid:     kid,
		signer:  js,
		options: opts,
	}, nil
}

func (s *rsaSigner) KeyID() string { return s.kid }

func (s *rsaSigner) Sign(info TaskInfo) (string, error) {
	now := time.Now().UTC()

	jti, err := randomJTI()
	if err != nil {
		return "", err
	}

	claims := TaskClaims{
		Issuer:    s.options.Issuer,
		Audience:  s.options.Audience,
		Subject:   fmt.Sprintf("task:%d", info.TaskID),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(s.options.TTL).Unix(),
		JWTID:     jti,

		TaskID:       info.TaskID,
		ProjectID:    info.ProjectID,
		ProjectName:  info.ProjectName,
		TemplateID:   info.TemplateID,
		TemplateName: info.TemplateName,
		UserID:       info.UserID,
		Username:     info.Username,
	}

	return josejwt.Signed(s.signer).Claims(claims).Serialize()
}

// JWKS returns the JWKS document containing the current public key.
func (s *rsaSigner) JWKS() ([]byte, error) {
	jwk := jose.JSONWebKey{
		Key:       &s.key.PublicKey,
		KeyID:     s.kid,
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}
	set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
	return json.Marshal(set)
}

func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("jwt: no PEM block found in key file")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("jwt: key is not RSA (got %T)", key)
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("jwt: unsupported PEM block type %q", block.Type)
	}
}

func computeKID(pub *rsa.PublicKey) (string, error) {
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
