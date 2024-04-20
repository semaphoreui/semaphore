package subscription

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

const rsaPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAn/...more RSA public key data...AB
-----END PUBLIC KEY-----`

// Function to parse RSA public key
func parseRSAPublicKeyFromPEM(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded public key: %v", err)
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("unknown type of public key")
	}
}

type Token struct {
	SubscriptionType string    `json:"subscription_type"`
	UserCount        int       `json:"user_count"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// ParseToken Function to verify a JWT
func ParseToken(tokenString string) (res Token, err error) {
	// Parse the RSA public key
	rsaPub, err := parseRSAPublicKeyFromPEM(rsaPublicKey)
	if err != nil {
		return
	}

	// Define a function to return the key for verification
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return rsaPub, nil
	}

	// Parse and verify the token
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return
	}
	if !token.Valid {
		err = fmt.Errorf("token is not valid")
		return
	}

	res.ExpiresAt, err = time.Parse("", token.Header["expires_at"].(string))
	res.SubscriptionType = token.Header["subscription_type"].(string)
	res.UserCount = token.Header["user_count"].(int)

	err = res.Validate()
	if err != nil {
		return
	}

	return
}

func (t *Token) Validate() error {
	return nil
}

func GetToken(store db.Store) (res Token, err error) {
	token, err := store.GetOption("subscription_token")
	if err != nil {
		return
	}

	if token == "" {
		err = fmt.Errorf("token is empty")
		return
	}

	res, err = ParseToken(token)
	if err != nil {
		return
	}

	return
}
