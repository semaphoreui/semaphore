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
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtNDwVQqLJsxiDffnZChd
sOSTITn3v2bQP03t6yZPDxBSgoV6sP/6eq20ZY2LpFURV7C+wpPVB3InTC0NXpm1
bWEGCUspWBxTI0YgPyPSQ57a2lFPPj/paqf9So4YjZeSJv0ozt9FukZ8GlozAJjl
RvunsbaYY1E/VClTCTp6wX2HUlQmhQ8cbWajreTagfM5S3V/wF6sWUiuiR5HFxlX
Ns4q8cYmyMCUVyDarxWJiXNDEUh0IHu1XZkm4zL1Lrgv857jY4sGdfqkShKYiym8
KnqMfaeeMSSvTenjSo32F0tXmTS+5gMONhDM17fb+a9hFmNlIaoJ9MrH6/hCvrbj
lQIDAQAB
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
	State     string    `json:"state"`
	Key       string    `json:"key"`
	Plan      string    `json:"plan"`
	Users     int       `json:"users"`
	ExpiresAt time.Time `json:"expiresAt"`
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

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = fmt.Errorf("failed to parse claims")
		return
	}

	res.ExpiresAt, err = time.Parse(time.RFC3339, claims["expiresAt"].(string))
	res.Plan = claims["plan"].(string)
	res.Key = claims["key"].(string)
	res.Users = int(claims["users"].(float64))

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
		err = db.ErrNotFound
		return
	}

	res, err = ParseToken(token)
	if err != nil {
		return
	}

	return
}

func HasActiveSubscription(store db.Store) bool {
	_, err := GetToken(store)

	if err != nil {
		return false
	}

	return true

	//return token.State == "active"
}
