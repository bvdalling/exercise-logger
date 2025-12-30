package utils

import (
	"crypto/rand"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 10
	recoverySecretLength = 32
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword compares a password with a hash
func ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRecoverySecret generates a cryptographically secure 32-character recovery secret
// Uses alphanumeric characters (A-Z, a-z, 0-9)
func GenerateRecoverySecret() (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	randomBytes := make([]byte, recoverySecretLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	secret := make([]byte, recoverySecretLength)
	for i := 0; i < recoverySecretLength; i++ {
		secret[i] = chars[randomBytes[i]%byte(len(chars))]
	}
	return string(secret), nil
}

// HashRecoverySecret hashes the recovery secret using bcrypt
func HashRecoverySecret(secret string) (string, error) {
	return HashPassword(secret)
}

// CompareRecoverySecret compares recovery secret with hash
func CompareRecoverySecret(secret, hash string) bool {
	return ComparePassword(secret, hash)
}

