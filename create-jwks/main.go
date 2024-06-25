package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/MicahParks/jwkset"
)

// write in PEM format
func writePrivateKey(filename string, privateKey *rsa.PrivateKey) error {
	privateKeyFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	return pem.Encode(privateKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
}

func writeJwks(filename string, KID string, privateKey *rsa.PrivateKey) error {
	jwk, err := jwkset.NewJWKFromKey(privateKey, jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{
			KID: KID,
			USE: "sig",
		},
	})
	if err != nil {
		return err
	}
	result, err := json.Marshal(jwkset.JWKSMarshal{
		Keys: []jwkset.JWKMarshal{jwk.Marshal()},
	})
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(result)
	return err
}

func main() {
	jwksOutputFilename := os.Args[1]
	privateKeyOutputFilename := os.Args[2]

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	err = writePrivateKey(privateKeyOutputFilename, privateKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("Private key", privateKeyOutputFilename, "(PEM format)")

	err = writeJwks(jwksOutputFilename, "test-key", privateKey)
	if err != nil {
		panic(err)
	}
	fmt.Println("JWKS written to", jwksOutputFilename)
}
