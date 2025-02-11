package main

import (
	"bytes"
	"compress/zlib"
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"sync"
	"time"
)

// CRLManager manages a Certificate Revocation List (CRL)
type CRLManager struct {
	serialRevocationMap map[string]pkix.RevokedCertificate
	mutex               sync.Mutex
}

// NewCRLManager initializes and returns a new CRLManager
func NewCRLManager() *CRLManager {
	return &CRLManager{
		serialRevocationMap: make(map[string]pkix.RevokedCertificate),
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

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf) // Create a new zlib writer

	_, err := writer.Write(data) // Write data to the zlib writer
	if err != nil {
		return nil, err
	}

	err = writer.Close() // Close to flush the compressed data
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil // Return compressed data
}

// uncompress decompresses zlib-compressed data.
func uncompress(compressedData []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(compressedData)) // Create a new zlib reader
	if err != nil {
		return nil, err
	}
	defer reader.Close() // Ensure the reader is closed after use

	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader) // Copy the decompressed data into buffer
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil // Return decompressed data
}

// GenerateCRL generates a PEM-encoded CRL using CreateRevocationList
func (cm *CRLManager) GenerateCRL(rand io.Reader, issuerCert *x509.Certificate, issuerKey crypto.Signer) ([]byte, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var revokedCertificates []pkix.RevokedCertificate
	for _, entry := range cm.serialRevocationMap {
		revokedCertificates = append(revokedCertificates, entry)
	}

	template := x509.RevocationList{
		Number:              big.NewInt(time.Now().Unix()),
		ThisUpdate:          time.Now(),
		NextUpdate:          time.Now().Add(365 * 24 * time.Hour),
		RevokedCertificates: revokedCertificates,
	}

	crlBytes, err := x509.CreateRevocationList(rand, &template, issuerCert, issuerKey)
	if err != nil {
		return nil, err
	}

	pemBlock := &pem.Block{
		Type:  "X509 CRL",
		Bytes: crlBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}

func loadCertificateAndKey(certPath, keyPath string) (*x509.Certificate, crypto.Signer, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	block, _ = pem.Decode(keyPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode key PEM")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func main() {
	for i := 0; i < 6; i++ {
		fmt.Println("--------")
		// Set the number of CRL entries
		n := int(math.Pow10(i))
		cm := NewCRLManager()

		// Recompute the bit-string capacities
		capacityByte := 20 * n
		percentageRevoked := float64(0.01)
		numberRevoked := int64(math.Round(float64(capacityByte*8) * percentageRevoked))
		fmt.Printf("[i] Status list capacity (%d bytes, %d bits)\n", capacityByte, capacityByte*8)
		fmt.Printf("[i] percentage of revoked %.2f bytes, number of revoked %d\n", percentageRevoked*100, numberRevoked)

		// Create a randomly populated bit string with a pre-defined number of revocations
		bytes, _ := generateRandomNBitValue(capacityByte, int(numberRevoked))
		h := sha256.New()
		h.Write(bytes)
		// fmt.Printf("[t] digest before compression: %x\n", h.Sum(nil))

		// Compress
		bytes, _ = compress(bytes)

		// Uncompress - testing purposes
		uncompressed, _ := uncompress(bytes)
		h = sha256.New()
		h.Write(uncompressed)
		// fmt.Printf("[t] digest after compression:  %x\n", h.Sum(nil))

		chunkSize := 20
		totalLen := len(bytes)
		n1 := totalLen / chunkSize

		for j := 0; j < n1; j++ { // Process full 20-byte chunks
			serial := new(big.Int).SetBytes(bytes[j*chunkSize : (j+1)*chunkSize])
			revocationTime := time.Now()
			reasonCode := 1 // Key Compromise
			cm.AddOrUpdateCRLEntry(serial, revocationTime, reasonCode)
		}

		// Handle remaining bytes (if not a multiple of 20)
		remaining := totalLen % chunkSize
		if remaining > 0 {
			lastChunk := make([]byte, chunkSize)       // Create a 20-byte buffer
			copy(lastChunk, bytes[n1*chunkSize:])      // Copy remaining bytes
			serial := new(big.Int).SetBytes(lastChunk) // Convert to big.Int
			revocationTime := time.Now()
			reasonCode := 1 // Key Compromise
			cm.AddOrUpdateCRLEntry(serial, revocationTime, reasonCode)
		}

		// Load issuer certificate and key
		issuerCert, issuerKey, err := loadCertificateAndKey("certs/rootCA.crt", "certs/rootCA.key")
		if err != nil {
			fmt.Println("Error loading CA certificate and key:", err)
			return
		}

		crl, err := cm.GenerateCRL(rand.Reader, issuerCert, issuerKey)
		if err != nil {
			fmt.Println("Error generating CRL:", err)
			return
		}

		fmt.Printf("Bit string capacity: %d, CRL size %d bytes\n", n*160, len(crl))

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
