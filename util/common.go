package util

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"regexp"
)

func ValidateUUIDv4(id string) bool {
	uuidV4Regex := regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidV4Regex.MatchString(id)
}

// Generate16ByteHex generates a random 16-byte hex string
func Generate16ByteHex() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatal("Failed to generate random bytes:", err)
	}
	return hex.EncodeToString(bytes)
}
