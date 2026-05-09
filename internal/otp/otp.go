package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	minCodeLength = 1
	maxCodeLength = 18
)

// GenerateCode generates a secure numeric OTP code with the requested length.
// Leading zeros are preserved. Lengths outside the supported range return an error.
func GenerateCode(length int) (string, error) {
	if length < minCodeLength || length > maxCodeLength {
		return "", fmt.Errorf("otp code length must be between %d and %d", minCodeLength, maxCodeLength)
	}

	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	num, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random OTP: %w", err)
	}

	return fmt.Sprintf("%0*d", length, num), nil
}

// Generate6DigitCode generates a secure 6-digit numeric OTP code.
// The code is always exactly 6 digits, with leading zeros preserved (e.g., "000123").
// Returns an error if random generation fails.
func Generate6DigitCode() (string, error) {
	return GenerateCode(6)
}
