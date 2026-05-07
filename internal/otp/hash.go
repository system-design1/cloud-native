package otp

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashCode returns a stable SHA-256 hash for an OTP code.
func HashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// VerifyCode compares a plaintext OTP code with a previously stored hash.
func VerifyCode(code string, storedHash string) bool {
	codeHash := HashCode(code)
	return subtle.ConstantTimeCompare([]byte(codeHash), []byte(storedHash)) == 1
}

// TODO: Implement  HMAC
/*
HashCode از SHA-256 و VerifyCode از constant-time compare استفاده کرده؛ برای Phase 1 قابل قبول است، ولی بعداً بهتر است salt/pepper یا HMAC اضافه کنیم.
*/
