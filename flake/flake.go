// Package flake manages IDs very similar to snowflake IDs used by Twitter and others.
package flake

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

var ErrInvalid = errors.New("invalid flake")

const zeros = "0"

var zerob = []byte(zeros)

type ID uint64

func (id ID) IsZero() bool {
	return id == 0
}

func (id ID) String() string {
	if id == 0 {
		return zeros
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(id))
	return hex.EncodeToString(b[:])
}

func (id ID) StringBytes(quote byte) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(id))
	if quote != 0 {
		s := make([]byte, 18)
		s[0] = quote
		s[17] = quote
		hex.Encode(s[1:17], b[:])
		return s
	} else {
		s := make([]byte, 16)
		hex.Encode(s, b[:])
		return s
	}
}

func (v ID) EncodeMsgpack(enc *msgpack.Encoder) error {
	if v == 0 {
		return enc.EncodeNil()
	} else {
		return enc.EncodeUint64(uint64(v))
	}
}
func (v *ID) DecodeMsgpack(dec *msgpack.Decoder) error {
	code, err := dec.PeekCode()
	if err != nil {
		return err
	}
	if code == msgpcode.Nil {
		*v = 0
		return nil
	} else if msgpcode.IsBin(code) || msgpcode.IsString(code) {
		s, err := dec.DecodeString()
		if err != nil {
			return err
		}
		*v, err = Parse(s)
		return err
	} else {
		n, err := dec.DecodeUint64()
		*v = ID(n)
		return err
	}
}

func Parse(s string) (ID, error) {
	return ParseBytes([]byte(s))
}

func ParseBytes(s []byte) (ID, error) {
	if bytes.Equal(s, zerob) {
		return 0, nil
	}
	if len(s) != 16 {
		return 0, ErrInvalid
	}
	var b [8]byte
	_, err := hex.Decode(b[:], s)
	if err != nil {
		return 0, ErrInvalid
	}
	return ID(binary.BigEndian.Uint64(b[:])), nil
}

func (id ID) MarshalText() ([]byte, error) {
	if id == 0 {
		return []byte(nil), nil
	}
	return id.StringBytes(0), nil
}

func (id *ID) UnmarshalText(b []byte) (err error) {
	if len(b) == 0 {
		*id = 0
		return nil
	}
	*id, err = ParseBytes(b)
	return
}

func (id *ID) Set(s string) (err error) {
	if s == "" {
		*id, err = 0, nil
	} else {
		*id, err = Parse(s)
	}
	return
}

// func (id ID) MarshalFlat() ([]byte, error) {
// 	return id.StringBytes(0), nil
// }

// func (id *ID) UnmarshalFlat(b []byte) (err error) {
// 	*id, err = ParseBytes(b)
// 	return
// }

func (id ID) MarshalJSON() ([]byte, error) {
	if id == 0 {
		return nullb, nil
	}
	return id.StringBytes('"'), nil
}

func (id *ID) UnmarshalJSON(input []byte) error {
	n := len(input)
	if n == 0 || bytes.Equal(input, nullb) {
		*id = 0
		return nil
	}
	if n != 18 || input[0] != '"' || input[17] != '"' {
		*id = 0
		return ErrInvalid
	}
	var err error
	*id, err = ParseBytes(input[1:17])
	return err
}

var nullb = []byte("null")

// Bubblehouse epoch starts on Jan 1, 2020 GMT
const EpochMs uint64 = 1577836800_000

const (
	timeBits              = 40
	nodeBits              = 8
	seqBits               = 16
	nodeShift             = seqBits
	timeShift             = (nodeBits + seqBits)
	timeMask       uint64 = (2 << timeBits) - 1
	nodeMask       uint64 = (2 << nodeBits) - 1
	seqMask        uint64 = (2 << seqBits) - 1
	nodeAndSeqMask uint64 = (2 << (nodeBits + seqBits)) - 1
)

func MinAt(tm time.Time) ID {
	return Build(MillisecondsFromTime(tm), 0, 0)
}

func Build(ms uint64, node uint64, seq uint64) ID {
	if node > nodeMask {
		panic(fmt.Sprintf("node value too large: %d", node))
	}
	if seq > seqMask {
		panic(fmt.Sprintf("seq value too large: %d", node))
	}
	return ID((ms << timeShift) | (node << nodeShift) | seq)
}

func (id ID) Milliseconds() uint64 {
	return uint64(id >> nodeShift)
}

func (id ID) Node() uint64 {
	return uint64(id>>nodeShift) & nodeMask
}

func (id ID) Seq() uint64 {
	return uint64(id) & seqMask
}

func (id ID) Time() time.Time {
	u := id.Milliseconds() + EpochMs
	s, ms := u/1000, u%1000
	return time.Unix(int64(s), int64(ms*uint64(time.Millisecond)))
}

func (id ID) MsFirst() ID {
	return id & ID(^nodeAndSeqMask)
}

func (id ID) MsLast() ID {
	return id | ID(nodeAndSeqMask)
}

func MillisecondsFromTime(tm time.Time) uint64 {
	v := int64(tm.UnixMilli()) - int64(EpochMs)
	if v < 0 || uint64(v) > timeMask {
		panic(fmt.Errorf("time %v is unrepresentable as bubble milliseconds", tm))
	}
	return uint64(v)
}

func FirstAt(tm time.Time) ID {
	return Build(MillisecondsFromTime(tm), 0, 0)
}

type Memento uint64

type Gen struct {
	node uint64

	lock    sync.Mutex
	lastMs  uint64
	lastSeq uint64
}

func NewGen(mem Memento, node uint64) *Gen {
	return &Gen{
		node:    node,
		lastMs:  uint64(mem),
		lastSeq: seqMask,
	}
}

func (g *Gen) Memento() Memento {
	return Memento(g.lastMs)
}

func (g *Gen) New() ID {
	return g.NewAt(time.Now())
}

func (g *Gen) NewAt(now time.Time) ID {
	ms := MillisecondsFromTime(now)

	g.lock.Lock()
	defer g.lock.Unlock()
	if ms > g.lastMs {
		g.lastMs = ms
		g.lastSeq = 1
	} else {
		g.lastSeq++
		if g.lastSeq > seqMask {
			g.lastMs++
			g.lastSeq = 1
		}
	}
	return Build(g.lastMs, g.node, g.lastSeq)
}
