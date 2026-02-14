package util

import (
	"crypto/rand"
	"math/big"
)

// Base62 characters (0-9, A-Z, a-z)
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GenerateShortCode generates a random base62 string of specified length
// This creates short, URL-safe codes like: "aB3xK9Qm"
func GenerateShortCode(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		result[i] = base62Chars[num.Int64()]
	}
	return string(result)
}

// GenerateCallLink creates a short, clean call link
// Format: vcme.com/join/aB3xK9Qm (8 characters by default)
func GenerateCallLink(appDomain string) string {
	shortCode := GenerateShortCode(8)
	return appDomain + "/join/" + shortCode
}

// GenerateCallCode generates just the code without domain
// Useful for displaying to users: "Meeting Code: aB3xK9Qm"
func GenerateCallCode() string {
	return GenerateShortCode(8)
}
