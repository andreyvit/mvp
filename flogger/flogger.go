// Package flogger stands for Fire Logger; keep your naughty thoughts to yourself.
package flogger

import (
	"bytes"
	"fmt"
)

type Context interface {
	Logf(format string, args ...any)
	AppendLogPrefix(buf *bytes.Buffer)
	AppendLogSuffix(buf *bytes.Buffer)
}

var DefaultContext = panicOutput{}

func Log(lc Context, message string, args ...any) {
	if lc == nil {
		lc = DefaultContext
	}

	plen := prefixLen(message)
	prefix, message := message[:plen], message[plen:]

	var buf bytes.Buffer
	if len(prefix) > 0 {
		buf.WriteString(prefix)
	}
	if lc != nil {
		lc.AppendLogPrefix(&buf)
	}
	fmt.Fprintf(&buf, message, args...)
	if lc != nil {
		lc.AppendLogSuffix(&buf)
	}
	lc.Logf("%s", buf.Bytes())
}

var prefixes = []string{"ERROR: ", "WARNING: ", "NOTICE: "}

func prefixLen(message string) int {
	ml := len(message)
	for _, p := range prefixes {
		pl := len(p)
		if ml >= pl && message[:pl] == p {
			return pl
		}
	}
	return 0
}

type prefixCtx struct {
	Context
	prefix string
}

func Prefix(parent Context, prefix string) Context {
	return &prefixCtx{parent, prefix}
}

func (lc *prefixCtx) AppendLogPrefix(buf *bytes.Buffer) {
	lc.Context.AppendLogPrefix(buf)
	buf.WriteString(lc.prefix)
}

type funcOutput func(format string, args ...any)

func ToFunc(f func(format string, args ...any)) Context {
	return funcOutput(f)
}

func (lc funcOutput) Logf(format string, args ...any) {
	lc(format, args)
}

func (lc funcOutput) AppendLogPrefix(buf *bytes.Buffer) {
}

func (lc funcOutput) AppendLogSuffix(buf *bytes.Buffer) {
}

type panicOutput struct{}

func (_ panicOutput) Logf(format string, args ...any) {
	panic(fmt.Errorf("unhandled log: "+format, args...))
}

func (_ panicOutput) AppendLogPrefix(buf *bytes.Buffer) {
}

func (_ panicOutput) AppendLogSuffix(buf *bytes.Buffer) {
}
