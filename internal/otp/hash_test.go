package otp

import "testing"

func TestHashCodeDeterministic(t *testing.T) {
	code := "123456"

	first := HashCode(code)
	second := HashCode(code)

	if first != second {
		t.Fatalf("HashCode should return deterministic hashes: first=%q second=%q", first, second)
	}
}

func TestHashCodeDifferentCodes(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
	}{
		{name: "different numeric codes", a: "123456", b: "654321"},
		{name: "leading zero matters", a: "012345", b: "12345"},
		{name: "empty differs from code", a: "", b: "000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if HashCode(tt.a) == HashCode(tt.b) {
				t.Fatalf("HashCode(%q) should differ from HashCode(%q)", tt.a, tt.b)
			}
		})
	}
}

func TestVerifyCode(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		storedHash string
		want       bool
	}{
		{
			name:       "matching code and hash",
			code:       "123456",
			storedHash: HashCode("123456"),
			want:       true,
		},
		{
			name:       "non matching code and hash",
			code:       "654321",
			storedHash: HashCode("123456"),
			want:       false,
		},
		{
			name:       "malformed hash",
			code:       "123456",
			storedHash: "not-a-valid-sha256-hex-hash",
			want:       false,
		},
		{
			name:       "empty malformed hash",
			code:       "123456",
			storedHash: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyCode(tt.code, tt.storedHash); got != tt.want {
				t.Fatalf("VerifyCode(%q, %q) = %v, want %v", tt.code, tt.storedHash, got, tt.want)
			}
		})
	}
}
