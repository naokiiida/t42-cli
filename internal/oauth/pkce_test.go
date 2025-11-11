package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGeneratePKCEParams(t *testing.T) {
	params, err := GeneratePKCEParams()
	if err != nil {
		t.Fatalf("GeneratePKCEParams failed: %v", err)
	}

	// Test verifier length
	if len(params.CodeVerifier) < 43 || len(params.CodeVerifier) > 128 {
		t.Errorf("Code verifier length %d not in range [43, 128]", len(params.CodeVerifier))
	}

	// Test challenge is valid base64url
	_, err = base64.RawURLEncoding.DecodeString(params.CodeChallenge)
	if err != nil {
		t.Errorf("Code challenge is not valid base64url: %v", err)
	}

	// Test challenge is correct hash of verifier
	expectedHash := sha256.Sum256([]byte(params.CodeVerifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(expectedHash[:])
	if params.CodeChallenge != expectedChallenge {
		t.Errorf("Code challenge doesn't match hash of verifier")
	}
}

func TestCodeVerifierUniqueness(t *testing.T) {
	params1, _ := GeneratePKCEParams()
	params2, _ := GeneratePKCEParams()

	if params1.CodeVerifier == params2.CodeVerifier {
		t.Error("Generated same code verifier twice (should be cryptographically random)")
	}

	if params1.CodeChallenge == params2.CodeChallenge {
		t.Error("Generated same code challenge twice")
	}
}

func TestCodeChallengeFormat(t *testing.T) {
	params, _ := GeneratePKCEParams()

	// Challenge should be 43 characters (SHA256 hash, base64url encoded, no padding)
	if len(params.CodeChallenge) != 43 {
		t.Errorf("Code challenge length %d, expected 43", len(params.CodeChallenge))
	}

	// Should only contain base64url characters
	for _, c := range params.CodeChallenge {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			t.Errorf("Code challenge contains invalid character: %c", c)
		}
	}
}
