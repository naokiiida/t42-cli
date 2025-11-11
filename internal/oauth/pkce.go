package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const (
	// RFC 7636 specifies code_verifier length: 43-128 characters
	codeVerifierLength = 64 // bytes (will be 86-88 chars when base64url encoded)
)

// PKCEParams holds the PKCE code verifier and challenge
type PKCEParams struct {
	CodeVerifier  string
	CodeChallenge string
}

// GeneratePKCEParams generates a code verifier and corresponding code challenge
// according to RFC 7636
func GeneratePKCEParams() (*PKCEParams, error) {
	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	challenge := generateCodeChallenge(verifier)

	return &PKCEParams{
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
	}, nil
}

// generateCodeVerifier creates a cryptographically random code verifier
// RFC 7636 Section 4.1:
// code_verifier = high-entropy cryptographic random STRING using the
// unreserved characters [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
// with a minimum length of 43 characters and a maximum length of 128 characters.
func generateCodeVerifier() (string, error) {
	// Generate random bytes
	bytes := make([]byte, codeVerifierLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Base64url encode (no padding)
	// This produces a string with unreserved characters as required
	verifier := base64.RawURLEncoding.EncodeToString(bytes)

	return verifier, nil
}

// generateCodeChallenge creates the code challenge from the verifier
// RFC 7636 Section 4.2:
// code_challenge = BASE64URL(SHA256(ASCII(code_verifier)))
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return challenge
}
