package main

import "math"

func (cm *CRLManager) NewRandomBitStringExtension(capacityByte int, percentageRevoked float64, compress bool) ([]byte, error) {

	// Input params
	numberRevoked := int64(math.Round(float64(capacityByte) * 8 * percentageRevoked))

	// Generate a random sample
	bitStringListBytes, _ := generateRandomNBitValue(capacityByte, int(numberRevoked))

	// Compress the bit string
	if compress {
		var err error
		bitStringListBytes, err = Compress(bitStringListBytes)
		if err != nil {
			return nil, err
		}
	}

	// Create a BitString CRL with extension
	return cm.NewCRLBitStringExtension(bitStringListBytes)
}
