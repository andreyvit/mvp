package logging

import (
	"io"
	"strings"
)

type TB interface {
	Log(args ...any)
	Logf(format string, args ...any)
}

func TBWriter(t TB) io.Writer {
	return &slogCapturer{t}
}

type slogCapturer struct {
	t TB
}

func (c *slogCapturer) Write(buf []byte) (int, error) {
	msg := string(buf)
	origLen := len(msg)
	msg = strings.TrimSuffix(msg, "\n")
	c.t.Log(msg)
	return origLen, nil
}
