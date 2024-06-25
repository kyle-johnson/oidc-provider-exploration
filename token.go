package main

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Workspace string `json:"workspace,omitempty"`
	jwt.RegisteredClaims
}

type TokenIssuer struct {
	// need to get details like kid and use from here
	jwks jwkset.JWKSMarshal

	// these are fixed at runtime
	audience string
	issuer   string

	// required to generate tokens
	privateKey *rsa.PrivateKey
}

func NewTokenIssuer(jwks jwkset.JWKSMarshal, audiance string, issuer string, privateKey *rsa.PrivateKey) *TokenIssuer {
	return &TokenIssuer{
		jwks:       jwks,
		privateKey: privateKey,
		audience:   audiance,
		issuer:     issuer,
	}
}

// WARNING: this endpoint is not protected!
func (t *TokenIssuer) IssuerHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		Workspace: "moon", // this is a custom claim; note that AWS doesn't support custom claims
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "yoda",
			Audience:  []string{t.audience},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			Issuer:    t.issuer,
		},
	})
	token.Header["kid"] = t.jwks.Keys[0].KID

	tokenString, err := token.SignedString(t.privateKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(tokenString))
}
