package main

import (
	"crypto/sha256"
	"fmt"
	"math"
)

func benchmarkCRLBitString(cm *CRLManager) {

	for i := 0; i < 3; i++ {
		fmt.Println("--------")
		// Set the number of CRL entries
		n := int(math.Pow10(i))

		// Recompute the bit-string capacities
		capacityByte := 125 * n // we make it multiple of 125 so that bit list capacity is a pow. of 10
		percentageRevoked := float64(0.1)
		numberRevoked := int64(math.Round(float64(capacityByte*8) * percentageRevoked))

		fmt.Printf("[i] Status list capacity (%d bytes, %d bits)\n", capacityByte, capacityByte*8)
		fmt.Printf("[i] Percentage of revoked %.2f%%, number of revoked %d\n", percentageRevoked*100, numberRevoked)

		// Create a randomly populated bit string with a pre-defined number of revocations
		bytes, _ := generateRandomNBitValue(capacityByte, int(numberRevoked))
		// fmt.Println(hex.EncodeToString(bytes))
		h := sha256.New()
		h.Write(bytes)
		// fmt.Printf("[t] digest before compression: %x\n", h.Sum(nil))
		sizeUncompressed := len(bytes)

		// Compress
		bytes, _ = Compress(bytes)
		sizeCompressed := len(bytes)

		fmt.Printf("[c] Compression Ratio: %.2f\n", float64(sizeUncompressed)/float64(sizeCompressed))

		crl, err := cm.NewCRLBitStringExtension(bytes)
		if err != nil {
			fmt.Println("Error generating CRL:", err)
			return
		}

		fmt.Printf("[.] PEM CRL size %d bytes\n[r] Bit string capacity: %d\n", len(crl), capacityByte*8)
	}
}

func benchmarkOld(cm *CRLManager) {

	for i := 0; i < 6; i++ {
		fmt.Println("--------")
		// Set the number of CRL entries
		n := int(math.Pow10(i))

		// Recompute the bit-string capacities
		capacityByte := 20 * n
		percentageRevoked := float64(0.1)
		numberRevoked := int64(math.Round(float64(capacityByte*8) * percentageRevoked))
		// fmt.Printf("[i] Status list capacity (%d bytes, %d bits)\n", capacityByte, capacityByte*8)
		// fmt.Printf("[i] percentage of revoked %.2f bytes, number of revoked %d\n", percentageRevoked*100, numberRevoked)

		// Create a randomly populated bit string with a pre-defined number of revocations
		bytes, _ := generateRandomNBitValue(capacityByte, int(numberRevoked))
		h := sha256.New()
		h.Write(bytes)
		// fmt.Printf("[t] digest before compression: %x\n", h.Sum(nil))

		// Compress
		bytes, _ = Compress(bytes)

		// // Uncompress - testing purposes
		// uncompressed, _ := uncompress(bytes)
		// h = sha256.New()
		// h.Write(uncompressed)
		// // fmt.Printf("[t] digest after compression:  %x\n", h.Sum(nil))

		// chunkSize := 20
		// totalLen := len(bytes)
		// n1 := totalLen / chunkSize

		// for j := 0; j < n1; j++ { // Process full 20-byte chunks
		// 	serial := new(big.Int).SetBytes(bytes[j*chunkSize : (j+1)*chunkSize])
		// 	revocationTime := time.Now()
		// 	reasonCode := 1 // Key Compromise
		// 	cm.AddOrUpdateCRLEntry(serial, revocationTime, reasonCode)
		// }

		// // Handle remaining bytes (if not a multiple of 20)
		// remaining := totalLen % chunkSize
		// if remaining > 0 {
		// 	lastChunk := make([]byte, chunkSize)       // Create a 20-byte buffer
		// 	copy(lastChunk, bytes[n1*chunkSize:])      // Copy remaining bytes
		// 	serial := new(big.Int).SetBytes(lastChunk) // Convert to big.Int
		// 	revocationTime := time.Now()
		// 	reasonCode := 1 // Key Compromise
		// 	cm.AddOrUpdateCRLEntry(serial, revocationTime, reasonCode)
		// }

		// crl, err := cm.GenerateCRL(rand.Reader, issuerCert, issuerKey)
		// if err != nil {
		// 	fmt.Println("Error generating CRL:", err)
		// 	return
		// }
		crl, err := cm.NewCRLBitStringExtension(bytes)
		if err != nil {
			fmt.Println("Error generating CRL:", err)
			return
		}

		fmt.Printf("Bit string capacity: %d, CRL size %d bytes\n", n*160, len(crl))
	}
}
