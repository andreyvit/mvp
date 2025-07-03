package jwt

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
	"unsafe"
)

var (
	ErrCorrupted          = errors.New("token corrupted")
	ErrAlg                = errors.New("token uses a wrong algorithm")
	ErrExpired            = errors.New("token expired")
	ErrNotYetValid        = errors.New("token not valid yet")
	ErrTooLong            = errors.New("token too long")
	ErrSignature          = errors.New("token signature invalid")
	ErrSignatureCorrupted = errors.New("token signature corrupted")

	MaxTokenLen        = 8000 // MaxTokenLen is the safety limit to avoid decoding very long data
	ExpectedClaimCount = 10   // ExpectedClaimCount is a starting size for the claims map
)

type Algorithm string

const (
	TokenID     = "jti" // TokenID is a unique identifier for this token.
	Issuer      = "iss" // Issuer is the principal that issued the token
	Audience    = "aud" // Audience identifies the recipients the token is intended for
	Subject     = "sub" // Subject is the user/account/etc. that this token authorizes
	IssuedAt    = "iat" // IssuedAt is a Unix timestamp for when the token was issued
	ExpiresAt   = "exp" // ExpiresAt is a Unix timestamp for when the token should expire
	NotBeforeAt = "nbf" // NotBeforeAt is a timestamp this token should not be accepted until

	Alg   = "alg" // Alg is a header field identifying the signing algorithm
	Typ   = "typ" // Typ is a header field that must be set to "JWT"
	KeyID = "kid" // KeyID is a header field, an opaque string identifying the key used

	Forever time.Duration = 1<<63 - 1 // Forever is a token validity that never expires

	stackClaimsSpace     = 512
	hs256SignatureEncLen = 43                                     // RawURLEncoding.EncodedLen(sha256.Size)
	hs256Header          = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" // {"alg":"HS256","typ":"JWT"} => base64url
	rs256Header          = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9" // {"alg":"RS256","typ":"JWT"} => base64url
	jwtTyp               = "JWT"

	HS256          Algorithm = "HS256"
	HS512          Algorithm = "HS512"
	RS256          Algorithm = "RS256"
	MinHS256KeyLen           = 32
	MaxHS256KeyLen           = 64 // anything longer is hashed to 32 bytes
)

type Claims map[string]any

func New(subject string, validity time.Duration) Claims {
	return NewAt(subject, validity, time.Now())
}

func NewAt(subject string, validity time.Duration, now time.Time) Claims {
	if validity == 0 {
		// accepting 0 would allow a misconfiguration to escalate into a security issue
		panic("zero validity is invalid, use Forever for non-expiring tokens")
	}

	c := make(Claims, ExpectedClaimCount)
	c[IssuedAt] = now.Unix()
	if validity != Forever {
		c[ExpiresAt] = now.Add(validity).Unix()
	}
	if subject != "" {
		c[Subject] = subject
	}
	return c
}

func (c Claims) String(key string) string {
	if v, ok := c[key].(string); ok {
		return v
	} else {
		return ""
	}
}

func (c Claims) Int64(key string) (int64, bool) {
	switch v := c[key].(type) {
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return n, true
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

func (c Claims) Time(key string) time.Time {
	if v, ok := c.Int64(key); ok && v != 0 {
		return time.Unix(v, 0)
	} else {
		return time.Time{}
	}
}

func (c Claims) ExpiresAt() time.Time {
	return c.Time(ExpiresAt)
}

func (c Claims) Issuer() string {
	return c.String(Issuer)
}

func (c Claims) Subject() string {
	return c.String(Subject)
}

func (c Claims) TokenID() string {
	return c.String(TokenID)
}

func (c Claims) KeyID() string {
	return c.String(KeyID)
}

func (c Claims) ValidateTime(tolerance time.Duration) error {
	return c.ValidateTimeAt(tolerance, time.Now())
}

func (c Claims) ValidateTimeAt(tolerance time.Duration, now time.Time) error {
	if exp := c.ExpiresAt(); !exp.IsZero() {
		if now.After(exp.Add(tolerance)) {
			return ErrExpired
		}
	}
	if exp := c.Time(NotBeforeAt); !exp.IsZero() {
		if now.Before(exp.Add(-tolerance)) {
			return ErrNotYetValid
		}
	}
	return nil
}

type header struct {
	Alg   string `json:"alg"`
	Typ   string `json:"typ"`
	KeyID string `json:"kid"`
}

func Sign(claims, headerClaims Claims, signer Signer) ([]byte, error) {
	rawClaims, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	if headerClaims == nil {
		headerClaims = make(Claims)
	}
	headerClaims[Typ] = jwtTyp
	headerClaims[Alg] = signer.Algorithm()
	if kid := signer.KeyID(); kid != "" {
		headerClaims[KeyID] = kid
	}

	rawHeader, err := json.Marshal(headerClaims)
	if err != nil {
		return nil, err
	}

	encHeader := make([]byte, base64.RawURLEncoding.EncodedLen(len(rawHeader)))
	base64.RawURLEncoding.Encode(encHeader, rawHeader)

	encClaims := make([]byte, base64.RawURLEncoding.EncodedLen(len(rawClaims)))
	base64.RawURLEncoding.Encode(encClaims, rawClaims)

	dataToSign := []byte(string(encHeader) + "." + string(encClaims))

	sig, err := signer.Sign(dataToSign)
	if err != nil {
		return nil, err
	}

	encSig := make([]byte, base64.RawURLEncoding.EncodedLen(len(sig)))
	base64.RawURLEncoding.Encode(encSig, sig)

	token := []byte(string(dataToSign) + "." + string(encSig))
	return token, nil
}

func SignString(claims, headerClaims Claims, signer Signer) (string, error) {
	token, err := Sign(claims, headerClaims, signer)
	if err != nil {
		return "", err
	}
	return unsafe.String(&token[0], len(token)), nil
}

func SignHS256String(claims, headerClaims Claims, key []byte) string {
	token, err := SignString(claims, headerClaims, NewHS256Signer(key, claims.KeyID()))
	if err != nil {
		panic(err)
	}
	return token
}

func SignRS25String(claims, headerClaims Claims, key *rsa.PrivateKey) string {
	token, err := SignString(claims, headerClaims, NewRS256Signer(key, claims.KeyID()))
	if err != nil {
		panic(err)
	}
	return token
}

// Token is the result of parsing a JWT token.
type Token struct {
	claims     Claims
	alg        Algorithm
	keyID      string
	dataToSign []byte
	sig        []byte
}

func (t *Token) Claims() Claims {
	return t.claims
}

func (t *Token) Alg() Algorithm {
	return t.alg
}

func (t *Token) KeyID() string {
	if t.keyID != "" {
		return t.keyID
	}
	return t.claims.KeyID()
}

// ParseString decodes JWT parts of a token.
func ParseString(rawToken string) (Token, error) {
	var token Token
	err := token.ParseString(rawToken)
	return token, err
}

// Parse decodes JWT parts of a token.
func Parse(rawToken []byte) (Token, error) {
	var token Token
	err := token.Parse(rawToken)
	return token, err
}

// ValidateWith verifies that this token’s algorithm matches the verifier,
// then checks the signature with verifier.Verify().
func (t *Token) ValidateWith(verifier Verifier) error {
	if t.alg != verifier.Algorithm() {
		return ErrAlg
	}
	if base64.RawURLEncoding.DecodedLen(len(t.sig)) != verifier.SigLen() {
		return ErrSignatureCorrupted
	}
	rawSig := make([]byte, verifier.SigLen())
	n, err := base64.RawURLEncoding.Decode(rawSig, t.sig)
	if err != nil || n != len(rawSig) {
		return ErrSignatureCorrupted
	}

	err = verifier.Verify(t.dataToSign, rawSig[:n])
	if err != nil {
		return err // e.g. ErrSignature
	}
	return nil
}

func (token *Token) ValidateHS256(key []byte) error {
	return token.ValidateWith(NewHS256Verifier(key, token.KeyID()))
}

func (token *Token) ValidateRS256(key *rsa.PublicKey) error {
	return token.ValidateWith(NewRS256Verifier(key, token.KeyID()))
}

func DecodeStringAt(
	rawToken string,
	verifier Verifier,
	tolerance time.Duration,
	now time.Time,
) (Claims, error) {
	var token Token
	if err := token.ParseString(rawToken); err != nil {
		return nil, err
	}

	if err := token.ValidateWith(verifier); err != nil {
		return nil, err
	}

	claims := token.Claims()
	if err := claims.ValidateTimeAt(tolerance, now); err != nil {
		return nil, err
	}

	return claims, nil
}

func DecodeHS256String(rawToken string, tolerance time.Duration, key []byte) (Claims, error) {
	return DecodeHS256StringAt(rawToken, key, tolerance, time.Now())
}

func DecodeHS256StringAt(
	rawToken string,
	key []byte,
	tolerance time.Duration,
	now time.Time,
) (Claims, error) {
	return DecodeStringAt(rawToken, NewHS256Verifier(key, ""), tolerance, now)
}

func DecodeRS256String(rawToken string, tolerance time.Duration, pubKey *rsa.PublicKey) (Claims, error) {
	return DecodeRS256StringAt(rawToken, pubKey, tolerance, time.Now())
}

func DecodeRS256StringAt(
	rawToken string,
	pubKey *rsa.PublicKey,
	tolerance time.Duration,
	now time.Time,
) (Claims, error) {
	return DecodeStringAt(rawToken, NewRS256Verifier(pubKey, ""), tolerance, now)
}

// ParseString decodes JWT parts of a token.
func (token *Token) ParseString(rawToken string) error {
	return token.Parse(unsafe.Slice(unsafe.StringData(rawToken), len(rawToken)))
}

// Parse decodes JWT parts of a token.
func (token *Token) Parse(rawToken []byte) error {
	if len(rawToken) > MaxTokenLen {
		return ErrTooLong
	}

	i1 := bytes.IndexByte(rawToken, '.')
	if i1 < 0 {
		return ErrCorrupted
	}
	i2 := bytes.IndexByte(rawToken[i1+1:], '.')
	if i2 < 0 {
		return ErrCorrupted
	}
	i2 += i1 + 1

	// Header part
	h := rawToken[:i1]
	if string(h) == hs256Header {
		token.alg = HS256
	} else if string(h) == rs256Header {
		token.alg = RS256
	} else {
		dbuf := make([]byte, base64.RawURLEncoding.DecodedLen(len(h)))
		n, err := base64.RawURLEncoding.Decode(dbuf, h)
		if err != nil {
			return ErrCorrupted
		}
		var hdr header
		err = json.Unmarshal(dbuf[:n], &hdr)
		if err != nil {
			return ErrCorrupted
		}
		if hdr.Typ != jwtTyp {
			return ErrCorrupted
		}
		token.alg = Algorithm(hdr.Alg)
		token.keyID = hdr.KeyID
	}
	token.sig = rawToken[i2+1:]
	token.dataToSign = rawToken[:i2]

	// Claims part
	c := make(Claims, ExpectedClaimCount)
	{
		raw := rawToken[i1+1 : i2]
		n := base64.RawURLEncoding.DecodedLen(len(raw))

		// if claims data is small enough, decode into a stack buffer to avoid allocation
		var stackBuf [stackClaimsSpace]byte
		var buf []byte
		if n < cap(stackBuf) {
			buf = stackBuf[:]
		} else {
			buf = make([]byte, n)
		}
		n, err := base64.RawURLEncoding.Decode(buf, raw)
		if err != nil {
			return ErrCorrupted
		}
		dec := json.NewDecoder(bytes.NewReader(buf[:n]))
		dec.UseNumber()
		err = dec.Decode(&c)
		if err != nil {
			return ErrCorrupted
		}
	}
	token.claims = c
	return nil
}
