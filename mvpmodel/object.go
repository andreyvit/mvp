package mvpm

import (
	"fmt"
	"strings"

	"github.com/andreyvit/mvp/flake"
)

type Object interface {
	FlakeID() flake.ID
	ObjectType() Type
}

type Type int

const (
	TypeNone        = Type(0)
	TypeUser        = Type(1)
	TypeFirstCustom = Type(100)
)

var (
	typeStrings   = make(map[Type]string)
	typesByString = make(map[string]Type)
)

func RegisterType(at Type, names ...string) {
	if s, ok := typeStrings[at]; ok && s != names[0] {
		panic(fmt.Errorf("Type %d already registered as %q, cannot register it as %q", at, s, names[0]))
	}
	typeStrings[at] = names[0]
	for _, name := range names {
		if cnflct, ok := typesByString[name]; ok && cnflct != at {
			panic(fmt.Errorf("Type string %q already registered as %d, cannot register it as %d", name, cnflct, at))
		}
		typesByString[name] = at
	}
}

func (v Type) String() string {
	return typeStrings[v]
}

func ParseType(s string) (Type, error) {
	if at, ok := typesByString[s]; ok {
		return at, nil
	} else {
		return TypeNone, fmt.Errorf("invalid actor type %q", s)
	}
}

func (v Type) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Type) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseType(string(b))
	return err
}

type Ref struct {
	Type Type
	ID   flake.ID
}

func (ar Ref) IsZero() bool {
	return ar.Type == TypeNone
}

func (ar Ref) String() string {
	return fmt.Sprintf("%v:%v", ar.Type, ar.ID)
}

func RefTo(obj Object) Ref {
	return Ref{obj.ObjectType(), obj.FlakeID()}
}

func ParseRef(s string) (Ref, error) {
	if s == "" {
		return Ref{}, nil
	}
	typeStr, idStr, ok := strings.Cut(s, ":")
	if ok {
		typ, err := ParseType(typeStr)
		if err == nil || typeStr == "" {
			id, err := flake.Parse(idStr)
			if err == nil {
				return Ref{typ, id}, nil
			}
		}
	}
	return Ref{}, fmt.Errorf("invalid ref %q", s)
}

func (v Ref) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Ref) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseRef(string(b))
	return err
}
