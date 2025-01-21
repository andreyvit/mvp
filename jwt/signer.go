package jwt

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

type Signer interface {
	Algorithm() Algorithm
	KeyID() string
	// Sign takes the combined "header.claims" (already base64-encoded) and returns the raw binary signature.
	// The caller is responsible for base64-encoding that signature.
	Sign(data []byte) ([]byte, error)
}

// ----------------------------------------------------------------------------
// HS256 Signer
// ----------------------------------------------------------------------------

// HS256Signer signs using HMAC-SHA256 with a symmetric key.
type HS256Signer struct {
	key   []byte
	keyID string
}

// NewHS256Signer returns an HS256 signer using the given key and optional keyID.
func NewHS256Signer(key []byte, keyID string) *HS256Signer {
	if len(key) == 0 {
		panic("HS256Signer: key is empty")
	}
	return &HS256Signer{key: key, keyID: keyID}
}

func (s *HS256Signer) Algorithm() Algorithm {
	return HS256
}

func (s *HS256Signer) KeyID() string {
	return s.keyID
}

// Sign uses HMAC-SHA256 on 'data'.
func (s *HS256Signer) Sign(data []byte) ([]byte, error) {
	h := hmac.New(sha256.New, s.key)
	_, _ = h.Write(data) // error is always nil
	return h.Sum(nil), nil
}

// ----------------------------------------------------------------------------
// RS256 Signer
// ----------------------------------------------------------------------------

type RS256Signer struct {
	privKey *rsa.PrivateKey
	keyID   string
}

// NewRS256Signer returns an RS256 signer using the given RSA private key.
func NewRS256Signer(privKey *rsa.PrivateKey, keyID string) *RS256Signer {
	if privKey == nil {
		panic("RS256Signer: private key is nil")
	}
	return &RS256Signer{privKey: privKey, keyID: keyID}
}

func (s *RS256Signer) Algorithm() Algorithm {
	return RS256
}

func (s *RS256Signer) KeyID() string {
	return s.keyID
}

// Sign uses RSA-SHA256 on 'data'.
func (s *RS256Signer) Sign(data []byte) ([]byte, error) {
	h := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, s.privKey, crypto.SHA256, h[:])
}
