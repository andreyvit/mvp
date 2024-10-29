package mvpdebug

import (
	"path"
	"runtime"
	"strconv"
	"unsafe"
)

func BriefStack(skip int) string {
	var pcbuf [10]uintptr
	n := runtime.Callers(2+skip, pcbuf[:])
	if n == 0 {
		return ""
	}
	pc := pcbuf[:n]
	frames := runtime.CallersFrames(pc)
	buf := make([]byte, 0, 1024)
	for {
		frame, more := frames.Next()

		if len(buf) > 0 {
			buf = append(buf, ' ')
		}

		if frame.Function != "" {
			buf = append(buf, frame.Function...)
		} else if frame.File != "" {
			base := path.Base(frame.File)
			dir := path.Base(path.Dir(frame.File))
			if dir != "" && dir != "." {
				buf = append(buf, dir...)
				buf = append(buf, '/')
			}
			buf = append(buf, base...)
			if frame.Line != 0 {
				buf = append(buf, ':')
				buf = strconv.AppendInt(buf, int64(frame.Line), 10)
			}
		} else {
			buf = append(buf, "<unknown>"...)
		}

		if !more {
			break
		}
	}

	return unsafe.String(&buf[0], len(buf))
}
