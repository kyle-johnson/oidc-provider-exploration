package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func loadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("PEM decoding failed")
	}
	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key type, expected 'RSA PRIVATE KEY' but got: %s", block.Type)
	}

	// Parse the private key
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
