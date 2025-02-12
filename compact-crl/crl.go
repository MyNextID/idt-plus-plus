package main

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// CRLManager manages a Certificate Revocation List (CRL)
type CRLManager struct {
	serialRevocationMap map[string]pkix.RevokedCertificate
	mutex               sync.Mutex
	key                 crypto.Signer    // CRL signing key
	crt                 x509.Certificate // CRL signing certificate
}

// NewCRLManager initializes and returns a new CRLManager
func NewCRLManager(_keyPath, _crtPath *string) *CRLManager {
	keyPath := "certs/rootCA.key"
	crtPath := "certs/rootCA.crt"
	if _keyPath != nil {
		keyPath = *_keyPath
	}
	if _crtPath != nil {
		crtPath = *_crtPath
	}
	crt, key, err := loadCertificateAndKey(crtPath, keyPath)
	if err != nil {
		panic(err)
	}
	return &CRLManager{
		serialRevocationMap: make(map[string]pkix.RevokedCertificate),
		key:                 key,
		crt:                 *crt,
	}
}

// AddOrUpdateCRLEntry adds or updates a revoked certificate entry in the CRLManager
func (cm *CRLManager) AddOrUpdateCRLEntry(serialNumber *big.Int, revocationTime time.Time, reasonCode int) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	serialStr := serialNumber.String()
	cm.serialRevocationMap[serialStr] = pkix.RevokedCertificate{
		SerialNumber: serialNumber,
		// RevocationTime: revocationTime,
		// Note: extensions don't make sense for compact CRL
		// Extensions: []pkix.Extension{
		// 	{
		// 		Id:    []int{2, 5, 29, 21}, // Reason Code extension OID
		// 		Value: []byte{byte(reasonCode)},
		// 	},
		// },
	}
}

// generateRandomNBitValue sets N random bits to 1 in a 160-bit value
func generateRandomNBitValue(size, n int) ([]byte, error) {
	var value = make([]byte, size)

	if n < 0 || n > size*8 {
		return value, fmt.Errorf("invalid number of bits: %d (must be between 0 and ...)", n)
	}

	setBits := make(map[int]bool) // Track already set bits to avoid duplicates

	for len(setBits) < n {
		randIndex, err := rand.Int(rand.Reader, big.NewInt(int64(size*8))) // Get a random bit position (0-159)
		if err != nil {
			return value, err
		}

		bitIndex := int(randIndex.Int64())
		if setBits[bitIndex] {
			continue // Skip if already set
		}

		setBits[bitIndex] = true
		byteIndex := bitIndex / 8
		bitOffset := bitIndex % 8
		value[byteIndex] |= (1 << bitOffset) // Set the bit to 1
	}

	return value, nil
}

// NewCRL generates a PEM-encoded CRL using a list of revoked serial numbers
func (cm *CRLManager) NewCRL(revokedSerialNumbers []big.Int) ([]byte, error) {

	// SN -> Revoked certs
	revokedCerts := []pkix.RevokedCertificate{}
	rTime := time.Now()
	for _, v := range revokedSerialNumbers {
		crt := pkix.RevokedCertificate{
			SerialNumber:   &v,
			RevocationTime: rTime,
		}
		revokedCerts = append(revokedCerts, crt)
	}

	// Construct a crl
	crl := x509.RevocationList{
		Number:              big.NewInt(time.Now().Unix()),
		ThisUpdate:          rTime,
		NextUpdate:          rTime.Add(365 * 24 * time.Hour),
		RevokedCertificates: revokedCerts,
	}

	// DER encoding
	crlBytes, err := x509.CreateRevocationList(rand.Reader, &crl, &cm.crt, cm.key)
	if err != nil {
		return nil, err
	}

	// Encode to PEM
	pemBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: crlBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}

// Create a new CRL with bit string extension
func (cm *CRLManager) NewCRLBitStringExtension(bitString []byte) ([]byte, error) {

	// Construct a crl
	crl := x509.RevocationList{
		Number:          big.NewInt(time.Now().Unix()),
		ThisUpdate:      time.Now(),
		NextUpdate:      time.Now().Add(365 * 24 * time.Hour),
		ExtraExtensions: []pkix.Extension{},
	}

	// Define a private extension
	bitStringExtension := pkix.Extension{
		Id:       []int{1, 2, 3, 4, 5}, // Private OID - for bit string
		Critical: false,
		Value:    bitString,
	}

	// Include the private extension
	crl.ExtraExtensions = append(crl.Extensions, bitStringExtension)

	// Create a CRL
	crlBytes, err := x509.CreateRevocationList(rand.Reader, &crl, &cm.crt, cm.key)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("[i] DER CRL size %d bytes\n", len(crlBytes))

	// Encode to PEM
	pemBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: crlBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}
