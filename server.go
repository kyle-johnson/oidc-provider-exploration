package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/MicahParks/jwkset"
)

type OIDCConfiguration struct {
	Issuer                           string   `json:"issuer"`
	JwksURI                          string   `json:"jwks_uri"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                  []string `json:"claims_supported"`
}

func createMux(prefix string, jwks []byte, oidcConfiguration OIDCConfiguration, issuerHandlerFunc http.HandlerFunc) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	// serve up the public key; notably this includes the kid and use
	mux.HandleFunc(fmt.Sprintf("%sjwks", prefix), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwks)
	})

	// openid-configuration
	cachedMarshalledConfiguration, err := json.Marshal(oidcConfiguration)
	if err != nil {
		return nil, err
	}
	mux.HandleFunc(fmt.Sprintf("%s.well-known/openid-configuration", prefix), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedMarshalledConfiguration)
	})

	// intentionally not prefixed
	mux.HandleFunc("/issue-token", issuerHandlerFunc)

	return mux, nil
}

func main() {
	jwksFilename := os.Args[1]
	privateKeyFilename := os.Args[2]

	baseUrl, ok := os.LookupEnv("BASE_URL")
	if !ok {
		panic("BASE_URL not set")
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		panic("PORT not set")
	}

	aud, ok := os.LookupEnv("AUD")
	if !ok {
		panic("AUD not set")
	}

	// may be empty
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/"
	}

	jwksBytes, err := os.ReadFile(jwksFilename)
	if err != nil {
		panic(err)
	}

	jwks := jwkset.JWKSMarshal{}
	err = json.Unmarshal(jwksBytes, &jwks)
	if err != nil {
		panic(err)
	}

	privateKey, err := loadPrivateKey(privateKeyFilename)
	if err != nil {
		panic(err)
	}

	tokenIssuer := NewTokenIssuer(jwks, aud, baseUrl, privateKey)

	fmt.Println("BASE_URL:", baseUrl)
	mux, err := createMux(prefix, jwksBytes, OIDCConfiguration{
		Issuer:                           baseUrl,
		JwksURI:                          fmt.Sprintf("%sjwks", baseUrl),
		ResponseTypesSupported:           []string{"id_token"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		ClaimsSupported: []string{"sub",
			"aud",
			"exp",
			"iat",
			"iss",
			"workspace"},
	}, tokenIssuer.IssuerHandlerFunc)
	if err != nil {
		panic(err)
	}
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%s", port), mux)
}
