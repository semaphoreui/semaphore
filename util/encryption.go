package util

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
)

func GeneratePrivateKey(privateKeyFile io.Writer) (publicKey string, err error) {
	// 1. Generate RSA Private Key (2048 bits)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	// 2. Encode the private key to PKCS#1 ASN.1 PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPem := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// 3. Write private key to file
	if err = pem.Encode(privateKeyFile, privateKeyPem); err != nil {
		return
	}

	publicKeyBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	publicKeyPem := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	var b bytes.Buffer
	publicKeyFile := bufio.NewWriter(&b)

	if err = pem.Encode(publicKeyFile, publicKeyPem); err != nil {
		return
	}

	publicKeyFile.Flush()

	publicKey = string(b.Bytes())
	return
}
