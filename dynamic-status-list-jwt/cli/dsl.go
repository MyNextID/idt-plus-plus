package main

import (
	"crypto"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

const dSLPeriod = float64(60) // dSL time period in seconds

// Generates revocation metadata and creates a revocation entry
func (s *Server) NewDslEntry(in string, detached bool) error {
	// we can revoke an IDT that has status information
	// or we can create a detached revocation token

	// JWT MUST have a jti
	// Load the jwt from the file and check if jti claim is present
	// Read the file
	data, err := os.ReadFile(in)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal the JSON
	var jwtData JWTData
	if err := json.Unmarshal(data, &jwtData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Parse the JWT
	idt, err := jwt.Parse([]byte(jwtData.Jwt), jwt.WithVerify(false))
	if err != nil {
		return err
	}

	// Check that the JWT contains jti claim
	var jti string
	err = idt.Get(jwt.JwtIDKey, &jti)
	if err != nil {
		return err
	}

	// Create dsl private metadata as jwt
	// Set: seed: base64url encoded random 16 byte value
	jtiDigest := sha256.Sum256([]byte(jti))
	seed := sha256.Sum256(append(s.Secret, jtiDigest[:]...))
	seedHex := hex.EncodeToString(seed[:])

	// Set the claims
	dslPrivateMetadata := jwt.New()
	dslPrivateMetadata.Set(jwt.SubjectKey, jti)
	dslPrivateMetadata.Set("seed", seedHex)
	signedDslPM, err := s.SignJWT(dslPrivateMetadata)

	// Add the jti to the list and set it to "valid"
	(*s.Dsl)[jti] = true
	if err != nil {
		return err
	}

	signedDetached := []byte{}
	if detached {
		t := jwt.New()
		// Set the sub claim
		t.Set(jwt.SubjectKey, jti)
		// Set the dSL distribution point is /dcp/list identifier
		// TODO: load this one via a variable
		t.Set("sdb", "http://localhost:PORT/sdb/1")
		signedDetached, err = s.SignJWT(t)
		if err != nil {
			return err
		}
	}

	// Save the result
	err = SaveJSON(JWTData{Jwt: jwtData.Jwt, PrivateMetadata: string(signedDslPM), DetachedDsl: string(signedDetached)}, in)
	if err != nil {
		return err
	}

	// Save the DSL to a file
	err = SaveDslMap("dsl-map.json", *s.Dsl)
	if err != nil {
		return err
	}
	// Recompute
	return s.RecomputeDslJwt()
}

// Create or load a new Dsl
func (s *Server) NewDsl(filename string) error {
	// We need a key-value map, key: jti, value: valid/invalid
	// Note: additional status metadata can be added (encrypted) -> for the v2 of the open source release
	// Note: we support a single map in the open source release

	// Try to load the dsl
	dslMap, err := LoadDslMap(filename)
	if err == nil {
		// map loaded, exit
		s.Dsl = &dslMap
		return nil
	}

	// Create a new empty map
	dslMap = make(map[string]bool)
	err = SaveDslMap(filename, dslMap)
	if err != nil {
		return err
	}

	s.Dsl = &dslMap

	return nil
}

// Save the dsl map to a file
func SaveDslMap(filename string, m map[string]bool) error {
	jsonData, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = os.WriteFile(filename, jsonData, 0600)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

type DslJWT struct {
	DslJwt string `json:"dsl_jwt"`
	Nbf    int64  `json:"nbf"`
}

// Recomputes the dSL every period
// Note: In production, three states should always be available: now-period, now, now+period
func (s *Server) DslService(period time.Duration) {
	// Create a new ticker with the specified period
	ticker := time.NewTicker(period)
	defer ticker.Stop() // Ensure the ticker is stopped when done

	// Infinite loop to listen for tick events
	for {
		select {
		case <-ticker.C: // The ticker sends a message every period
			s.RecomputeDslJwt() // Call your function
		}
	}
}

// Save the dsl as JWT
func (s *Server) RecomputeDslJwt() error {

	// Get the current time
	tNow := time.Now().Unix()
	return s.RecomputeDslJwtAt(tNow)
}

// Save the dsl as JWT
func (s *Server) RecomputeDslJwtAt(tNow int64) error {

	// Get the current time
	tNext := tNow + int64(dSLPeriod)

	// Compute the revocation identifiers
	sid := s.ComputeRevocationIdentifiers(s.Dsl, tNow)

	t := jwt.New()
	t.Set("typ", "dsl/v1")
	jwkThumbprint, err := s.PublicKey.Thumbprint(crypto.SHA256)
	if err != nil {
		return err
	}
	t.Set("iss", hex.EncodeToString(jwkThumbprint))
	t.Set(jwt.NotBeforeKey, tNow)
	t.Set(jwt.ExpirationKey, tNext-1)
	t.Set("nxt", tNext)
	t.Set("sid", sid) // revoked identifiers

	// Sign the jwt
	signed, err := s.SignJWT(t)
	if err != nil {
		return err
	}
	s.DslJwt = signed

	// Save the JWT
	return SaveJSON(DslJWT{DslJwt: string(signed), Nbf: tNow}, "dsl.json")
}

// Compute the revocation identifiers
func (s *Server) ComputeRevocationIdentifiers(m *map[string]bool, tNow int64) []string {

	// We store the results into the revocation list
	// Note: more space-efficient methods can be used, such as Bloom filter, CRLite, etc.
	revocationList := []string{}

	// Loop over the revocation statuses and compute the identifiers
	for jti, valid := range *m {

		jtiDigest := sha256.Sum256([]byte(jti))
		// Note: we selected this function for efficiency purposes; other seed derivation approaches can be used
		seed := sha256.Sum256(append(s.Secret, jtiDigest[:]...))

		// Compute the revocation entry
		reB64 := ComputeRevocationIdentifier(jti, seed[:], tNow, valid)

		revocationList = append(revocationList, reB64)

	}
	// Shuffle the elements
	rand.Shuffle(len(revocationList), func(i, j int) {
		revocationList[i], revocationList[j] = revocationList[j], revocationList[i]
	})

	return revocationList
}

func LoadDslMap(filename string) (map[string]bool, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var myMap map[string]bool
	if err := json.Unmarshal(data, &myMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return myMap, nil
}

// Revoke a credential
func (s *Server) Revoke(jti string) error {
	_, ok := (*s.Dsl)[jti]
	// If the key exists
	if !ok {
		return errors.New("jti not found. Create a new entry, first using the 'new' command")
	}
	// Update the state
	(*s.Dsl)[jti] = false

	// Recompute the DSL
	return s.RecomputeDslJwt()
}

func ComputeRevocationIdentifier(jti string, seed []byte, tNow int64, valid bool) string {

	jtiDigest := sha256.Sum256([]byte(jti))
	// Note: we selected this function for efficiency purposes; other seed derivation approaches can be used

	token, err := NewToken(seed, tNow)
	if err != nil {
		return ""
	}

	// valid = H(token, s_id)
	h256 := sha256.New()
	data := append(token, jtiDigest[:]...) // Store appended data properly
	h256.Write(data)
	revocationEntry := h256.Sum(nil)
	if !valid {
		// token is revoked
		h256 = sha256.New()
		h256.Write(revocationEntry)
		revocationEntry = h256.Sum(nil)
	}
	//
	reB64 := base64.RawURLEncoding.EncodeToString(revocationEntry)
	return reB64
}

func NewToken(seed []byte, tNow int64) ([]byte, error) {

	// t' = floor(t_now / period)
	t := uint64(math.Floor(float64(tNow) / dSLPeriod))

	// token = HMAC(seed, t’)
	tBytes := make([]byte, 8) // uint64 needs 8 bytes
	binary.BigEndian.PutUint64(tBytes, t)
	h := hmac.New(sha256.New, seed[:])
	_, err := h.Write(tBytes)
	if err != nil {
		return nil, err
	}

	token := h.Sum(nil)
	return token, nil

}

func ComputeRevocationIdentifierWithToken(jti string, tokenB64 string, valid bool) (string, error) {

	jtiDigest := sha256.Sum256([]byte(jti))
	// Note: we selected this function for efficiency purposes; other seed derivation approaches can be used

	token, err := base64.RawURLEncoding.DecodeString(tokenB64)
	if err != nil {
		return "", err
	}

	// valid = H(token, s_id)
	h256 := sha256.New()
	data := append(token, jtiDigest[:]...) // Store appended data properly
	h256.Write(data)
	revocationEntry := h256.Sum(nil)
	if !valid {
		// token is revoked
		h256 = sha256.New()
		h256.Write(revocationEntry)
		revocationEntry = h256.Sum(nil)
	}
	//
	reB64 := base64.RawURLEncoding.EncodeToString(revocationEntry)
	return reB64, nil
}
