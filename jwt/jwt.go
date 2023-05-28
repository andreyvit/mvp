package jwt

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
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
	Audience    = "aud" // Audience identifies the recipents the token is intended for
	Subject     = "sub" // Subject is the user/account /etc that this token authorizes access to
	IssuedAt    = "iat" // IssuedAt is a Unix timestamp for when the token was issued
	ExpiresAt   = "exp" // ExpiresAt is a Unix timestamp for when the token should expire
	NotBeforeAt = "nbf" // NotBeforeAt is a timestamp this token should not be accepted until

	Alg   = "alg" // Alg is a header field identifying the signing algorithm
	Typ   = "typ" // Typ is a header field that must be set to "JWT"
	KeyID = "kid" // KeyID is a header field, an opaque string identifying the key used

	Forever time.Duration = 1<<63 - 1 // Forever is validity duration of tokens that do not expire

	stackClaimsSpace     = 512
	hs256SignatureEncLen = 43                                     // RawURLEncoding.EncodedLen(sha256.Size)
	hs256Header          = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9" // {"alg":"HS256","typ":"JWT"}
	jwtTyp               = "JWT"

	HS256 Algorithm = "HS256"

	MinHS256KeyLen = 32
	MaxHS256KeyLen = 64 // anything longer is hashed to 32 bytes
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

func SignHS256String(claims, headerClaims Claims, key []byte) string {
	b := SignHS256(claims, headerClaims, key, nil)
	return unsafe.String(&b[0], len(b))
}

// SignHS256 produces a signed JWT token from the given claims.
func SignHS256(claims, headerClaims Claims, key []byte, buf []byte) []byte {
	if len(key) == 0 {
		panic("missing key")
	}
	rawClaims, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}
	var rawHeader []byte
	if headerClaims != nil {
		headerClaims[Typ] = jwtTyp
		headerClaims[Alg] = string(HS256)
		rawHeader, err = json.Marshal(headerClaims)
		if err != nil {
			panic(err)
		}
	}
	return SignHS256Raw(rawClaims, rawHeader, key, buf)
}

// SignHS256Raw produces a signed JWT token from the given raw claims.
func SignHS256Raw(claims, header []byte, key []byte, buf []byte) []byte {
	headerLen := len(header)
	if headerLen == 0 {
		headerLen = len(hs256Header)
	}

	claimLen := base64.RawURLEncoding.EncodedLen(len(claims))
	tokenLen := headerLen + 1 + claimLen + 1 + hs256SignatureEncLen

	if len(buf) < tokenLen {
		buf = make([]byte, tokenLen)
	}

	if len(header) == 0 {
		copy(buf, hs256Header)
	} else {
		copy(buf, header)
	}
	buf[headerLen] = '.'
	base64.RawURLEncoding.Encode(buf[headerLen+1:], claims)

	var hash [sha256.Size]byte
	alg := hmac.New(sha256.New, key)
	alg.Write(buf[:headerLen+1+claimLen])
	alg.Sum(hash[:0])

	buf[headerLen+1+claimLen] = '.'
	base64.RawURLEncoding.Encode(buf[headerLen+1+claimLen+1:], hash[:])
	return buf
}

// Token is a result of parsing a JWT token.
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
	return t.claims.String(KeyID)
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

func DecodeHS256String(rawToken string, tolerance time.Duration, key []byte) (Claims, error) {
	return DecodeHS256StringAt(rawToken, key, tolerance, time.Now())
}

func DecodeHS256StringAt(rawToken string, key []byte, tolerance time.Duration, now time.Time) (Claims, error) {
	var token Token
	err := token.ParseString(rawToken)
	if err != nil {
		return nil, err
	}
	err = token.ValidateHS256(key)
	if err != nil {
		return nil, err
	}
	c := token.Claims()
	err = c.ValidateTimeAt(tolerance, now)
	return c, err
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

	h := rawToken[:i1]
	if string(h) == hs256Header {
		token.alg = HS256
	} else {
		var hdr header
		err := json.Unmarshal(h, &hdr)
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

		// log.Printf("RawToken = %q", raw)

		n, err := base64.RawURLEncoding.Decode(buf, raw)
		if err != nil {
			return ErrCorrupted
		}

		// log.Printf("JSONToken = %s", buf[:n])

		dec := json.NewDecoder(bytes.NewReader(buf[:n]))
		dec.UseNumber()
		err = dec.Decode(&c)
		if err != nil {
			// log.Printf("JSON err: %v", err)
			return ErrCorrupted
		}
	}
	token.claims = c
	return nil
}

func (token *Token) Validate(alg Algorithm, key []byte) error {
	switch alg {
	case HS256:
		return token.ValidateHS256(key)
	default:
		panic("unsupported algorithm")
	}
}

func (token *Token) ValidateHS256(key []byte) error {
	if len(key) == 0 {
		panic("missing key")
	}
	if token.alg != HS256 {
		return ErrAlg
	}
	var actualHash, expectedHash [sha256.Size]byte
	if base64.RawURLEncoding.DecodedLen(len(token.sig)) != len(actualHash) {
		// log.Printf("base64.RawURLEncoding.DecodedLen(len(token.sig)) %d != len(actualHash) %d", base64.RawURLEncoding.DecodedLen(len(token.sig)), len(actualHash))
		return ErrSignatureCorrupted
	}
	n, err := base64.RawURLEncoding.Decode(actualHash[:], token.sig)
	if err != nil || n != len(actualHash) {
		return ErrSignatureCorrupted
	}

	alg := hmac.New(sha256.New, key)
	alg.Write(token.dataToSign)
	alg.Sum(expectedHash[:0])

	// log.Printf("StringToSign = %q", token[:i2])
	// log.Printf("expectedHash = %q", base64.RawURLEncoding.EncodeToString(expectedHash[:]))
	// log.Printf("actualHash = %q", base64.RawURLEncoding.EncodeToString(actualHash[:]))

	if 1 != subtle.ConstantTimeCompare(actualHash[:], expectedHash[:]) {
		return ErrSignature
	}
	return nil
}
