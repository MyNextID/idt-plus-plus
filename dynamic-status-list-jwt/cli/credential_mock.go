package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	scheme    = "http"
	host      = "localhost"
	port      = "4321"
	byteLen   = 16
	statusURL = "http://localhost:PORT/sdb/1"
)

// JWTData structure holds the JWT and associated metadata
type JWTData struct {
	Jwt             string `json:"jwt"`
	PrivateMetadata string `json:"private_metadata"`
	DetachedDsl     string `json:"detached_dsl_jwt"`
}

// IssueJWT generates a mock JWT with a unique ID (jti) and stores it at the specified path
func (s *Server) IssueJWT(out string) ([]byte, *string) {
	// Generate a random JTI (JWT ID)
	jtiByte := make([]byte, byteLen)
	if _, err := rand.Read(jtiByte); err != nil {
		fmt.Printf("[ERROR] failed to generate JTI: %v\n", err)
		return nil, nil
	}
	jti := hex.EncodeToString(jtiByte)

	// Create the JWT with claims
	tok := jwt.New()
	tok.Set(jwt.SubjectKey, "Alice")
	tok.Set(jwt.JwtIDKey, jti)
	tok.Set("sdb", statusURL)

	// Sign the JWT
	signedJWT, err := s.SignJWT(tok)
	if err != nil {
		fmt.Printf("[ERROR] failed to sign JWT: %v\n", err)
		return nil, nil
	}

	// Store the JWT in the specified file
	if err := SaveJSON(JWTData{Jwt: string(signedJWT)}, out); err != nil {
		fmt.Printf("[ERROR] failed to save JWT: %v\n", err)
		return nil, nil
	}

	return signedJWT, &jti
}

// Print reads and pretty-prints a JSON file
func Print(in string) {
	// Read the file
	data, err := os.ReadFile(in)
	if err != nil {
		fmt.Printf("[ERROR] failed to read file %s: %v\n", in, err)
		return
	}

	// Validate and unmarshal the JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		fmt.Printf("[ERROR] failed to unmarshal JSON: %v\n", err)
		return
	}

	// Pretty-print the formatted JSON
	formattedJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Printf("[ERROR] failed to format JSON: %v\n", err)
		return
	}

	fmt.Println(string(formattedJSON))
}

// NewSecret generates a secret using HMAC-SHA256 with the given key and JTI
func NewSecret(key, jtiByte []byte) []byte {
	h := hmac.New(sha256.New, key)
	return h.Sum(jtiByte)
}
