package jwt

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
)

type Verifier interface {
	SigLen() int
	Algorithm() Algorithm
	KeyID() string
	// Verify takes the "header.claims" data (already base64-encoded) and the raw decoded signature bytes.
	// It should return nil if verification succeeds, or a non-nil error otherwise.
	Verify(data, signature []byte) error
}

// ----------------------------------------------------------------------------
// HS256 Verifier
// ----------------------------------------------------------------------------

type HS256Verifier struct {
	key   []byte
	keyID string
}

// NewHS256Verifier constructs an HS256 verifier with the given secret `key` and optional `keyID`.
func NewHS256Verifier(key []byte, keyID string) *HS256Verifier {
	if len(key) == 0 {
		panic("HS256Verifier: key cannot be empty")
	}
	return &HS256Verifier{key: key, keyID: keyID}
}

func (v *HS256Verifier) Algorithm() Algorithm {
	return HS256
}

func (v *HS256Verifier) KeyID() string {
	return v.keyID
}

func (v *HS256Verifier) SigLen() int {
	return sha256.Size
}

// Verify checks that `signature` (raw bytes) matches the HMAC-SHA256 of `data`.
func (v *HS256Verifier) Verify(data, signature []byte) error {
	// Recompute the HMAC-SHA256 of data
	mac := hmac.New(sha256.New, v.key)
	mac.Write(data)
	expected := mac.Sum(nil)

	// Compare the provided signature to what we expect
	if subtle.ConstantTimeCompare(signature, expected) != 1 {
		return ErrSignature
	}
	return nil
}

// ----------------------------------------------------------------------------
// RS256 Verifier
// ----------------------------------------------------------------------------

type RS256Verifier struct {
	pubKey *rsa.PublicKey
	keyID  string
}

// NewRS256Verifier constructs an RS256 verifier with the given `pubKey` and optional `keyID`.
func NewRS256Verifier(pubKey *rsa.PublicKey, keyID string) *RS256Verifier {
	if pubKey == nil {
		panic("RS256Verifier: public key cannot be nil")
	}
	return &RS256Verifier{pubKey: pubKey, keyID: keyID}
}

func (v *RS256Verifier) Algorithm() Algorithm {
	return RS256
}

func (v *RS256Verifier) KeyID() string {
	return v.keyID
}

func (v *RS256Verifier) SigLen() int {
	return v.pubKey.Size()
}

// Verify checks that `signature` (raw bytes) matches the RSA-SHA256 of `data`.
func (v *RS256Verifier) Verify(data, signature []byte) error {
	h := sha256.Sum256(data)
	err := rsa.VerifyPKCS1v15(v.pubKey, crypto.SHA256, h[:], signature)
	if err != nil {
		return ErrSignature
	}
	return nil
}
