package mvpmetrics

import (
	"bytes"
	"io"
	"strconv"
)

type Writer struct {
	w      io.Writer
	buf    bytes.Buffer
	prefix string
}

func (mw *Writer) Reset(w io.Writer) {
	mw.w = w
	mw.buf.Reset()
}

func (mw *Writer) Flush() error {
	b := mw.buf.Bytes()
	if len(b) == 0 {
		return nil
	}
	_, err := mw.w.Write(b)
	return err
}

func (mw *Writer) SetPrefix(prefix string) string {
	old := mw.prefix
	mw.prefix = prefix
	return old
}

func (mw *Writer) WriteHeader(name string, typ, help string) {
	mw.buf.WriteString("# TYPE ")
	mw.buf.WriteString(name)
	mw.buf.WriteByte(' ')
	mw.buf.WriteString(typ)
	mw.buf.WriteByte('\n')
	mw.buf.WriteString("# HELP ")
	mw.buf.WriteString(name)
	mw.buf.WriteByte(' ')
	mw.buf.WriteString(help)
	mw.buf.WriteByte('\n')
}

func (mw *Writer) WriteFloat(name string, labelNames []string, labelValues []string, value float64) {
	mw.startMetric(name, labelNames, labelValues)
	var buf [64]byte // https://stackoverflow.com/q/1701055 says 24 is enough, but just in case
	mw.buf.Write(strconv.AppendFloat(buf[:0], value, 'f', -1, 64))
	mw.endMetric()
}

func (mw *Writer) WriteInt(name string, labelNames []string, labelValues []string, value int64) {
	mw.startMetric(name, labelNames, labelValues)
	var buf [20]byte // -9223372036854775808
	mw.buf.Write(strconv.AppendInt(buf[:0], value, 10))
	mw.endMetric()
}

func (mw *Writer) WriteUint(name string, labelNames []string, labelValues []string, value uint64) {
	mw.startMetric(name, labelNames, labelValues)
	var buf [20]byte // 18446744073709551615
	mw.buf.Write(strconv.AppendUint(buf[:0], value, 10))
	mw.endMetric()
}

func (mw *Writer) startMetric(name string, labelNames []string, labelValues []string) {
	mw.buf.WriteString(name)
	if len(labelNames) > 0 {
		mw.buf.WriteByte('{')
		for i, ln := range labelNames {
			lv := labelValues[i]
			if i > 0 {
				mw.buf.WriteByte(',')
			}
			mw.buf.WriteString(ln)
			mw.buf.WriteString(`="`)
			mw.buf.WriteString(lv)
			mw.buf.WriteByte('"')
		}
		mw.buf.WriteByte('}')
	}
	mw.buf.WriteByte(' ')
}

func (mw *Writer) endMetric() {
	mw.buf.WriteByte('\n')
}
