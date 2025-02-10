package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

// SaveJSON saves the variable into a JSON file
func SaveJSON(variable interface{}, path string) error {

	// Save to config.json
	data, err := json.MarshalIndent(variable, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to save : %w", err)
	}
	return nil
}

// LoadJSON loads JSON into a variable
func LoadJSON(variable interface{}, path string) error {
	// Check if config.json exists
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	// File exists, try to load the key
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(file, variable); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	return nil
}

// LoadJWK loads a JWK from a JSON file
func LoadJWK(path string) (jwk.Key, error) {
	// Check if config.json exists
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Read the file
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the JWK using jwk.ParseKey()
	key, err := jwk.ParseKey(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWK: %w", err)
	}

	return key, nil
}

type JWTContainer struct {
	JWT string `json:"jwt"`
}

func decodeSegment(seg string) (string, error) {
	seg = strings.TrimRight(seg, "=")
	decoded, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func prettyPrintJSON(jsonStr string) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(jsonStr), "", "  ")
	if err != nil {
		return jsonStr
	}
	return prettyJSON.String()
}

// Decode and print JWT
func PrintJWT(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	content := strings.TrimSpace(string(data))

	if strings.HasPrefix(content, "{") {
		var container JWTContainer
		if err := json.Unmarshal(data, &container); err != nil {
			log.Fatalf("Failed to parse JSON: %v", err)
		}
		content = container.JWT
	}

	parts := strings.Split(content, ".")
	if len(parts) != 3 {
		log.Fatalf("Invalid JWT format")
	}

	header, err := decodeSegment(parts[0])
	if err != nil {
		log.Fatalf("Failed to decode JWT header: %v", err)
	}
	payload, err := decodeSegment(parts[1])
	if err != nil {
		log.Fatalf("Failed to decode JWT payload: %v", err)
	}
	signature := parts[2]

	fmt.Println("Header:")
	fmt.Println(prettyPrintJSON(header))
	fmt.Println("Payload:")
	fmt.Println(prettyPrintJSON(payload))
	fmt.Println("Signature:")
	fmt.Printf("%s\n", signature)
}
