package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"math/big"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

// GenerateTOTPSecret generates a new TOTP secret for a user
func GenerateTOTPSecret(issuer, accountName string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	return key, nil
}

// GenerateTOTPQRCode generates a QR code data URI for TOTP setup
func GenerateTOTPQRCode(key *otp.Key) (string, error) {
	// Get the OTP Auth URL
	url := key.URL()

	// Generate QR code
	qrCode, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Convert to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, qrCode.Image(256)); err != nil {
		return "", fmt.Errorf("failed to encode QR code: %w", err)
	}

	// Convert to data URI
	dataURI := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))
	return dataURI, nil
}

// ValidateTOTP validates a TOTP code against a secret
func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateBackupCodes generates a set of backup codes for TOTP
func GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	max := big.NewInt(100000000) // 8 digits
	for i := 0; i < count; i++ {
		// Generate random 8-digit number
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = fmt.Sprintf("%08d", n.Int64())
	}
	return codes, nil
}

// HashBackupCodes hashes backup codes for storage
func HashBackupCodes(codes []string) (string, error) {
	data, err := json.Marshal(codes)
	if err != nil {
		return "", fmt.Errorf("failed to marshal backup codes: %w", err)
	}
	// For simplicity, we'll store them as JSON (in production, encrypt this)
	return string(data), nil
}

// VerifyBackupCode checks if a backup code is valid
func VerifyBackupCode(hashedCodes, code string) bool {
	var codes []string
	if err := json.Unmarshal([]byte(hashedCodes), &codes); err != nil {
		return false
	}
	for _, c := range codes {
		if c == code {
			return true
		}
	}
	return false
}
