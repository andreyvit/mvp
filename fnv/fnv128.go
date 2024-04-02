package fnv

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"

	"github.com/vmihailenco/msgpack/v5"
)

type Hash128 [2]uint64

const (
	Size128 = 16

	// prime128 is 2^88 + (2^8 + 0x3b)
	prime128Lower = 0x13b
	prime128Shift = 24 // 88 - 64
)

func String128(s string) Hash128 {
	sum := New128()
	sum.WriteString(s)
	return sum
}

func Bytes128(v []byte) Hash128 {
	sum := New128()
	sum.WriteBytes(v)
	return sum
}

func New128() Hash128 {
	return Hash128{0x6c62272e07bb0142, 0x62b821756295c58d}
}

func (s Hash128) String() string {
	var buf [2 * Size128]byte
	s.PutHex(buf[:])
	return string(buf[:])
}

func (s *Hash128) LoadBinary(buf []byte) {
	s[0] = binary.BigEndian.Uint64(buf[:8])
	s[1] = binary.BigEndian.Uint64(buf[8:16])
}

func (s *Hash128) PutBinary(buf []byte) {
	binary.BigEndian.PutUint64(buf[:8], s[0])
	binary.BigEndian.PutUint64(buf[8:16], s[1])
}

func (s *Hash128) PutHex(buf []byte) {
	var n [Size128]byte
	s.PutBinary(n[:])
	hex.Encode(buf, n[:])
}

func (s *Hash128) writeByte(v byte) {
	s[1] ^= uint64(v)
	s.WriteZero()
}

func (s *Hash128) WriteZero() {
	s0, s1 := bits.Mul64(prime128Lower, s[1])
	s0 += s[1]<<prime128Shift + prime128Lower*s[0]
	s[0], s[1] = s0, s1
}

func (hash *Hash128) WriteByte(v byte) error {
	hash.writeByte(v)
	return nil
}

// Common part

func (s *Hash128) WriteUint16(v uint16) {
	s.writeByte(byte(v >> 8))
	s.writeByte(byte(v))
}

func (s *Hash128) WriteUint32(v uint32) {
	s.writeByte(byte(v >> 24))
	s.writeByte(byte(v >> 16))
	s.writeByte(byte(v >> 8))
	s.writeByte(byte(v))
}

func (s *Hash128) WriteUint64(v uint64) {
	s.writeByte(byte(v >> 56))
	s.writeByte(byte(v >> 48))
	s.writeByte(byte(v >> 40))
	s.writeByte(byte(v >> 32))
	s.writeByte(byte(v >> 24))
	s.writeByte(byte(v >> 16))
	s.writeByte(byte(v >> 8))
	s.writeByte(byte(v))
}

func (s *Hash128) WriteInt(v int) {
	s.WriteUint64(uint64(v))
}

func (s *Hash128) WriteBool(v bool) {
	if v {
		s.writeByte(byte(1))
	} else {
		s.WriteZero()
	}
}

func (s *Hash128) WriteBytes(data []byte) {
	for _, c := range data {
		s.writeByte(c)
	}
}

func (s *Hash128) WriteBytesZ(data []byte) {
	s.WriteBytes(data)
	s.WriteZero()
}

func (s *Hash128) WriteString(str string) {
	for _, b := range []byte(str) {
		s.writeByte(b)
	}
}

func (s *Hash128) WriteStringZ(str string) {
	s.WriteString(str)
	s.WriteZero()
}

func (s *Hash128) Downmix16() uint16 {
	return uint16(s[0]>>48) ^ uint16(s[0]>>32) ^ uint16(s[0]>>16) ^ uint16(s[0]) ^ uint16(s[1]>>48) ^ uint16(s[1]>>32) ^ uint16(s[1]>>16) ^ uint16(s[1])
}

func (s Hash128) Downmix32() uint32 {
	return uint32(s[0]>>32) ^ uint32(s[0]) ^ uint32(s[1]>>32) ^ uint32(s[1])
}

func (s Hash128) Downmix64() uint64 {
	return s[0] ^ s[1]
}

func (s *Hash128) MarshalBinary() ([]byte, error) {
	b := make([]byte, Size128)
	binary.BigEndian.PutUint64(b[:8], s[0])
	binary.BigEndian.PutUint64(b[8:16], s[1])
	return b, nil
}

func (s *Hash128) UnmarshalBinary(b []byte) error {
	if len(b) != Size128 {
		return fmt.Errorf("fnv.Hash128: invalid size %d, wanted %d", len(b), Size128)
	}
	s.LoadBinary(b[:])
	return nil
}
func (v Hash128) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Hash128) UnmarshalText(b []byte) error {
	if len(b) != 2*Size128 {
		return fmt.Errorf("fnv.Hash128: invalid hex length %d, wanted %d", len(b), 2*Size128)
	}
	var buf [Size128]byte
	_, err := hex.Decode(buf[:], b)
	if err != nil {
		return err
	}
	v.LoadBinary(buf[:])
	return nil
}
func (v Hash128) EncodeMsgpack(enc *msgpack.Encoder) error {
	var buf [Size128]byte
	v.PutBinary(buf[:])
	enc.EncodeBytes(buf[:])
	return nil
}
func (v *Hash128) DecodeMsgpack(dec *msgpack.Decoder) error {
	b, err := dec.DecodeBytes()
	if err != nil {
		return err
	}
	v.LoadBinary(b)
	return nil
}
