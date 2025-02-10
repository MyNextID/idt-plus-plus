package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

// Derive a new status list identifier as a holder/wallet
func NewProof(in string, revoked bool, detached bool, timestamp int64) (*string, error) {

	// read the file
	data, err := os.ReadFile(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal the JSON
	var jwtData JWTData
	if err := json.Unmarshal(data, &jwtData); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	if jwtData.PrivateMetadata == "" {
		return nil, errors.New("private metadata missing")
	}

	t, err := jwt.Parse([]byte(jwtData.PrivateMetadata), jwt.WithVerify(false))
	if err != nil {
		return nil, err
	}

	var jti string
	err = t.Get(jwt.SubjectKey, &jti)
	if err != nil {
		return nil, err
	}

	var seedHex string
	err = t.Get("seed", &seedHex)
	if err != nil {
		return nil, err
	}
	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, err
	}

	// Get the current time
	tNow := time.Now().Unix()
	if timestamp != 0 {
		// if timestamp is provided, use it
		tNow = timestamp
	}

	// Compute the revocation identifiers
	reB64 := ComputeRevocationIdentifier(jti, seed, tNow, !revoked)
	token, err := NewToken(seed, tNow)
	if err != nil {
		return nil, err
	}
	tokenB64 := base64.RawURLEncoding.EncodeToString(token)

	err = SaveJSON(HolderProofPayload{Jti: jti, Token: tokenB64, Sid: reB64, Iat: tNow, Revoked: revoked}, "holder_status-list-identifier.json")
	if err != nil {
		return nil, err
	}
	return &reB64, nil

}

// Holder proof payload
type HolderProofPayload struct {
	Jti     string `json:"jti"`
	Token   string `json:"token"`
	Sid     string `json:"sid"`
	Iat     int64  `json:"iat"`
	Revoked bool   `json:"revoked"`
}
