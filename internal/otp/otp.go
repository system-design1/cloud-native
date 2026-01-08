package otp

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

// Generate6DigitCode generates a secure 6-digit numeric OTP code.
// The code is always exactly 6 digits, with leading zeros preserved (e.g., "000123").
// Returns an error if random generation fails.
func Generate6DigitCode() (string, error) {
	// Generate a random number in range [0, 999999]
	// We use rejection sampling to avoid modulo bias:
	// - 3 bytes give us range [0, 16777215]
	// - We reject values >= 16000000 to ensure uniform distribution
	// - This gives us 16,000,000 valid values, which is evenly divisible by 1,000,000
	const maxValid = 16000000
	const modulo = 1000000

	for {
		var buf [3]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return "", fmt.Errorf("failed to generate random OTP: %w", err)
		}

		// Convert bytes to uint32
		var numBytes [4]byte
		copy(numBytes[1:], buf[:])
		num := binary.BigEndian.Uint32(numBytes[:])

		// Reject values >= maxValid to ensure uniform distribution
		if num >= maxValid {
			continue
		}

		// Take modulo to get range [0, 999999]
		num = num % modulo

		// Format with leading zeros using %06d
		return fmt.Sprintf("%06d", num), nil
	}
}

