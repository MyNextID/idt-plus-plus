package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const serverConfig = "config.json"

// TODO: collect all the definitions and store them into variables

// Server variables
type Server struct {
	SecretKey jwk.Key
	PublicKey jwk.Key
	key       ecdsa.PrivateKey
	Secret    []byte           // sha256 hash of the secret key
	Dsl       *map[string]bool // true: valid, false: invalid/revoked
	DslJwt    []byte
}

// NewServer initializes and returns a new Server instance
func NewServer() *Server {
	// Load or create a server key
	key, err := getServerKey()
	if err != nil {
		fmt.Println("Error retrieving server key:", err)
		return nil
	}

	// Extract ECDSA private key from JWK key
	var sk = &ecdsa.PrivateKey{}
	err = jwk.Export(key, sk)
	if err != nil {
		fmt.Println("Error exporting private key:", err)
		return nil
	}

	// Convert the public key to JWK format
	pk, err := jwk.Import(sk.PublicKey)
	if err != nil {
		fmt.Println("Error importing public key:", err)
		return nil
	}

	// Initialize the Distributed Certificate Revocation List (DSL)
	dsl, err := LoadDslMap("dsl-map.json")
	if err != nil {
		fmt.Println("> Init a new dsl map")
		dsl = make(map[string]bool)
	}

	// Derive a secret from the private key (hashing the private key's D value)
	secret := sha256.New().Sum(sk.D.Bytes())

	// Return a new Server instance with initialized fields
	return &Server{
		SecretKey: key,      // Original server key
		PublicKey: pk,       // Public key in JWK format
		key:       *sk,      // ECDSA private key
		Secret:    secret,   // Derived secret
		Dsl:       &dsl,     // Distributed Certificate Revocation List
		DslJwt:    []byte{}, // JWT representation of the DSL (empty for now)
	}
}

// Load or Generate EC Private Key
func getServerKey() (jwk.Key, error) {
	key, err := LoadJWK(serverConfig)
	if err == nil {
		// Key loaded, exit
		return key, nil
	}

	// File doesn't exist, generate new ES256 key
	fmt.Println("Generating new EC key (ES256)")
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate EC key: %w", err)
	}

	// Convert to JWK format
	jwkKey, err := jwk.Import(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWK from private key: %w", err)
	}

	// Set key algorithm
	jwkKey.Set(jwk.AlgorithmKey, jwa.ES256)

	// Save to config.json
	err = SaveJSON(jwkKey, serverConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to store: %w", err)
	}

	return jwkKey, nil
}

// Sign a JWT with 'jwk' header claim
func (s *Server) SignJWT(t jwt.Token) ([]byte, error) {

	// Set the jwk header
	h := jws.NewHeaders()
	h.Set(jws.JWKKey, s.PublicKey)

	// Sign JWT
	return jwt.Sign(t, jwt.WithKey(jwa.ES256(), s.key, jws.WithProtectedHeaders(h)))
}

func Verify(dslJWTPath string, proofPath string) (bool, error) {

	var err error
	// Load the DSL
	var dsl DslJWT
	err = LoadJSON(&dsl, dslJWTPath)
	if err != nil {
		return false, err
	}
	t, err := jwt.Parse([]byte(dsl.DslJwt), jwt.WithVerify(false), jwt.WithValidate(false))
	if err != nil {
		return false, err
	}

	var rawSid []interface{}
	err = t.Get("sid", &rawSid)
	if err != nil {
		return false, err
	}

	// Convert []interface{} to []string
	var sid []string
	for _, v := range rawSid {
		if str, ok := v.(string); ok {
			sid = append(sid, str)
		} else {
			log.Println("Warning: sid contains a non-string value")
		}
	}

	// Load the DSL
	var h HolderProofPayload
	err = LoadJSON(&h, proofPath)
	if err != nil {
		return false, err
	}

	sidValid, err := ComputeRevocationIdentifierWithToken(h.Jti, h.Token, true)
	if err != nil {
		return false, err
	}
	for _, v := range sid {
		if v == sidValid {
			return false, nil
		}
	}
	sidInvalid, err := ComputeRevocationIdentifierWithToken(h.Jti, h.Token, false)
	if err != nil {
		return false, err
	}
	for _, v := range sid {
		if v == sidInvalid {
			return true, nil
		}
	}

	return false, errors.New("status list id not found")

}
