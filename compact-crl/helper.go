package main

import (
	"bytes"
	"compress/zlib"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// compress using zlib
func Compress(data []byte) ([]byte, error) {
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

// decompresses zlib-compressed data.
func Decompress(compressedData []byte) ([]byte, error) {
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

// load certificate and signing key
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

func ReadHexFile(filePath string) ([]byte, error) {
	// Read the file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Convert file content to string and remove whitespace
	hexString := string(data)

	// Decode the hex string to bytes
	decodedBytes, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %w", err)
	}

	return decodedBytes, nil
}
