package mvputil

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
)

type CallErrorCategory struct {
	Name string
}

func (err *CallErrorCategory) Error() string {
	return err.Name
}

type CallError struct {
	CallID            string
	IsNetwork         bool
	StatusCode        int
	Type              string
	Path              string
	Message           string
	RawResponseBody   []byte
	PrintResponseBody bool
	Cause             error
	Category1         *CallErrorCategory
	Category2         *CallErrorCategory
}

func (e *CallError) Error() string {
	return e.customError(true)
}
func (e *CallError) ShortError() string {
	return e.customError(false)
}
func (e *CallError) customError(withIdentity bool) string {
	var buf strings.Builder
	if withIdentity {
		if e.CallID != "" {
			buf.WriteString(e.CallID)
			buf.WriteString(": ")
		}
		if e.StatusCode != 0 {
			fmt.Fprintf(&buf, "HTTP %d", e.StatusCode)
		}
	}
	if e.IsNetwork {
		buf.WriteString("network: ")
	}
	if e.Type != "" {
		buf.WriteString(": ")
		buf.WriteString(e.Type)
	}
	if e.Category1 != nil {
		buf.WriteString(" <")
		buf.WriteString(e.Category1.Name)
		buf.WriteString(">")
	}
	if e.Category2 != nil {
		buf.WriteString(" <")
		buf.WriteString(e.Category2.Name)
		buf.WriteString(">")
	}
	if e.Message != "" {
		buf.WriteString(": ")
		buf.WriteString(e.Message)
	}
	if e.Path != "" {
		buf.WriteString(" [at ")
		buf.WriteString(e.Path)
		buf.WriteString("]")
	}
	if e.Cause != nil {
		buf.WriteString(": ")
		buf.WriteString(e.Cause.Error())
	}
	if e.PrintResponseBody {
		buf.WriteString("  // response: ")
		if len(e.RawResponseBody) == 0 {
			buf.WriteString("<empty>")
		} else if utf8.Valid(e.RawResponseBody) {
			buf.Write(e.RawResponseBody)
		} else {
			return fmt.Sprintf("<binary %d bytes>", len(e.RawResponseBody))
		}
	}
	return buf.String()
}

func (e *CallError) Unwrap() error {
	return e.Cause
}

func (e *CallError) IsUnprocessableEntity() bool {
	return e.StatusCode == http.StatusUnprocessableEntity
}

func (e *CallError) AddCategory(cat *CallErrorCategory) *CallError {
	if cat != nil && !e.IsInCategory(cat) {
		if e.Category1 == nil {
			e.Category1 = cat
		} else if e.Category2 == nil {
			e.Category2 = cat
		} else {
			panic("only 2 categories are supported per error")
		}
	}
	return e
}

func (e *CallError) Is(target error) bool {
	if cat, ok := target.(*CallErrorCategory); ok {
		return e.IsInCategory(cat)
	}
	return false
}

func (e *CallError) IsInCategory(cat *CallErrorCategory) bool {
	return cat != nil && (e.Category1 == cat || e.Category2 == cat)
}
