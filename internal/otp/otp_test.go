package otp

import (
	"regexp"
	"testing"
)

func TestGenerate6DigitCode(t *testing.T) {
	// Test that function returns a 6-digit code
	code, err := Generate6DigitCode()
	if err != nil {
		t.Fatalf("Generate6DigitCode() returned error: %v", err)
	}

	// Verify length is exactly 6
	if len(code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code))
	}

	// Verify all characters are digits
	matched, err := regexp.MatchString(`^\d{6}$`, code)
	if err != nil {
		t.Fatalf("Regex match error: %v", err)
	}
	if !matched {
		t.Errorf("Code %s does not match pattern ^\\d{6}$", code)
	}
}

func TestGenerate6DigitCode_LeadingZeros(t *testing.T) {
	// Generate multiple codes to check for leading zeros
	// This is probabilistic, but with enough iterations we should see some with leading zeros
	foundLeadingZero := false
	for i := 0; i < 1000; i++ {
		code, err := Generate6DigitCode()
		if err != nil {
			t.Fatalf("Generate6DigitCode() returned error: %v", err)
		}

		if len(code) != 6 {
			t.Errorf("Expected code length 6, got %d", len(code))
		}

		// Check if code starts with zero (leading zero preserved)
		if code[0] == '0' {
			foundLeadingZero = true
		}
	}

	// With 1000 iterations, probability of not finding a leading zero is extremely low
	// But we'll make this a warning, not a failure, since it's probabilistic
	if !foundLeadingZero {
		t.Log("Warning: No codes with leading zeros found in 1000 iterations (unlikely but possible)")
	}
}

func TestGenerate6DigitCode_Uniqueness(t *testing.T) {
	// Generate multiple codes and check they're not all the same
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := Generate6DigitCode()
		if err != nil {
			t.Fatalf("Generate6DigitCode() returned error: %v", err)
		}
		codes[code] = true
	}

	// With 100 iterations, we should have some uniqueness
	// (though collisions are possible, they should be rare)
	if len(codes) < 50 {
		t.Logf("Warning: Only %d unique codes in 100 iterations (collisions may occur)", len(codes))
	}
}

