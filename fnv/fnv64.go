// Package fnv implements mixing of various values using 64-bit FNV1a alg.
package fnv

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type Hash64 uint64

const (
	Size64         = 8
	prime64 Hash64 = 1099511628211
)

func String64(s string) Hash64 {
	sum := New64()
	sum.WriteString(s)
	return sum
}

func Bytes64(v []byte) Hash64 {
	sum := New64()
	sum.WriteBytes(v)
	return sum
}

func New64() Hash64 {
	return 14695981039346656037
}

func (s Hash64) String() string {
	var buf [2 * Size64]byte
	s.PutHex(buf[:])
	return string(buf[:])
}

func (s *Hash64) LoadBinary(buf []byte) {
	*s = Hash64(binary.BigEndian.Uint64(buf))
}

func (s *Hash64) PutBinary(buf []byte) {
	binary.BigEndian.PutUint64(buf, uint64(*s))
}

func (s *Hash64) PutHex(buf []byte) {
	var n [Size64]byte
	s.PutBinary(n[:])
	hex.Encode(buf, n[:])
}

func (hash *Hash64) WriteZero() {
	*hash *= prime64
}

func (hash *Hash64) writeByte(v byte) {
	*hash ^= Hash64(byte(v))
	hash.WriteZero()
}

func (hash *Hash64) WriteByte(v byte) error {
	hash.writeByte(v)
	return nil
}

// Common part

func (s *Hash64) WriteUint16(v uint16) {
	s.writeByte(byte(v >> 8))
	s.writeByte(byte(v))
}

func (s *Hash64) WriteUint32(v uint32) {
	s.writeByte(byte(v >> 24))
	s.writeByte(byte(v >> 16))
	s.writeByte(byte(v >> 8))
	s.writeByte(byte(v))
}

func (s *Hash64) WriteUint64(v uint64) {
	s.writeByte(byte(v >> 56))
	s.writeByte(byte(v >> 48))
	s.writeByte(byte(v >> 40))
	s.writeByte(byte(v >> 32))
	s.writeByte(byte(v >> 24))
	s.writeByte(byte(v >> 16))
	s.writeByte(byte(v >> 8))
	s.writeByte(byte(v))
}

func (s *Hash64) WriteInt(v int) {
	s.WriteUint64(uint64(v))
}

func (s *Hash64) WriteBool(v bool) {
	if v {
		s.writeByte(byte(1))
	} else {
		s.WriteZero()
	}
}

func (s *Hash64) WriteBytes(data []byte) {
	for _, c := range data {
		s.writeByte(c)
	}
}

func (s *Hash64) WriteBytesZ(data []byte) {
	s.WriteBytes(data)
	s.WriteZero()
}

func (s *Hash64) Write(v []byte) (n int, err error) {
	s.WriteBytes(v)
	return len(v), nil
}

func (s *Hash64) WriteString(str string) {
	for _, b := range []byte(str) {
		s.writeByte(b)
	}
}

func (s *Hash64) WriteStringZ(str string) {
	s.WriteString(str)
	s.WriteZero()
}

func (s Hash64) Downmix8() uint8 {
	return uint8(s>>56) ^ uint8(s>>48) ^ uint8(s>>40) ^ uint8(s>>32) ^ uint8(s>>24) ^ uint8(s>>16) ^ uint8(s>>8) ^ uint8(s)
}

func (s Hash64) Downmix16() uint16 {
	return uint16(s>>48) ^ uint16(s>>32) ^ uint16(s>>16) ^ uint16(s)
}

func (s Hash64) Downmix32() uint32 {
	return uint32(s>>32) ^ uint32(s)
}

func (s Hash64) Downmix64() uint64 {
	return uint64(s)
}

func (s Hash64) MarshalBinary() ([]byte, error) {
	b := make([]byte, Size64)
	binary.BigEndian.PutUint64(b[:8], uint64(s))
	return b, nil
}

func (s *Hash64) UnmarshalBinary(b []byte) error {
	if len(b) != Size64 {
		return fmt.Errorf("fnv.Hash64: invalid size %d, wanted %d", len(b), Size64)
	}
	s.LoadBinary(b)
	return nil
}
func (v Hash64) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Hash64) UnmarshalText(b []byte) error {
	if len(b) != 2*Size64 {
		return fmt.Errorf("fnv.Hash64: invalid hex length %d, wanted %d", len(b), 2*Size64)
	}
	var buf [Size64]byte
	_, err := hex.Decode(buf[:], b)
	if err != nil {
		return err
	}
	v.LoadBinary(buf[:])
	return nil
}
func (v Hash64) EncodeMsgpack(enc *msgpack.Encoder) error {
	enc.EncodeUint64(uint64(v))
	return nil
}
func (v *Hash64) DecodeMsgpack(dec *msgpack.Decoder) error {
	u, err := dec.DecodeUint64()
	if err != nil {
		return err
	}
	*v = Hash64(u)
	return nil
}
