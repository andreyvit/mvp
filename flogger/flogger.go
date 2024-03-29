// Package flogger stands for Fire Logger; keep your naughty thoughts to yourself.
package flogger

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

type Context interface {
	Logf(format string, args ...any)
	AppendLogPrefix(buf *bytes.Buffer)
	AppendLogSuffix(buf *bytes.Buffer)
}

var ContextKey = contextKeyType{}

type contextKeyType struct{}

func (contextKeyType) String() string {
	return "flogger.ContextKey"
}

func ContextFrom(ctx context.Context) Context {
	if v := ctx.Value(ContextKey); v != nil {
		return v.(Context)
	}
	return DefaultContext
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

func Logx(lc Context, level slog.Level, message string, args ...slog.Attr) {
	args = append(args, slog.String("level", level.String()))
	Log(lc, Format(message, args...))
}

func Format(msg string, attrs ...slog.Attr) string {
	if len(attrs) == 0 {
		return msg
	} else {
		buf := []byte(msg)
		buf = append(buf, " -"...)
		buf = Append(buf, attrs...)
		return string(buf)
	}
}

func Append(buf []byte, attrs ...slog.Attr) []byte {
	var prefix []byte
	return appendPrefix(buf, &prefix, attrs...)
}

func appendPrefix(buf []byte, prefix *[]byte, attrs ...slog.Attr) []byte {
	for _, attr := range attrs {
		if attr.Value.Kind() == slog.KindGroup {
			n := len(*prefix)
			*prefix = append(*prefix, attr.Key...)
			*prefix = append(*prefix, '.')
			buf = appendPrefix(buf, prefix, attr.Value.Group()...)
			*prefix = (*prefix)[:n]
		} else {
			buf = append(buf, ' ')
			buf = append(buf, (*prefix)...)
			buf = append(buf, attr.Key...)
			buf = append(buf, '=')
			s := attr.Value.String()
			if strings.IndexByte(s, ' ') < 0 && strings.IndexByte(s, '\n') < 0 && strings.IndexByte(s, '"') < 0 {
				buf = append(buf, s...)
			} else {
				buf = strconv.AppendQuote(buf, s)
			}
		}
	}
	return buf
}
