package sse

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

var (
	ErrHTTPStatusNon200   = errors.New("HTTP status code indicates failure")
	ErrInvalidContentType = errors.New("invalid Content-Type, expected text/event-stream")

	// ErrCloseEventStream is a marker error that can be returned by
	// stream handling func to indicate that the event stream should be
	// closed.
	ErrCloseEventStream = errors.New("close event stream")
)

func VerifyResponse(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ErrHTTPStatusNon200
	}
	ctype := resp.Header.Get("Content-Type")
	if ctype != ContentType {
		return ErrInvalidContentType
	}
	return nil
}

var (
	dataPrefixBytes  = []byte("data:")
	eventPrefixBytes = []byte("event:")
	idPrefixBytes    = []byte("id:")
	retryPrefixBytes = []byte("retry:")
	bomBytes         = []byte{0xEF, 0xBB, 0xBF}
	asciiDigits      = regexp.MustCompile(`^[0-9]+$`)
)

func ParseStream(r io.ReadCloser, maxSize int, f func(id, event string, data []byte) error, retryf func(ms uint64)) error {
	defer r.Close()

	scanner := bufio.NewScanner(r)
	scanner.Split(scanLinesWithLF)

	// Use a stack-allocated buffer if we can fit lines into it
	var data [512]byte
	scanner.Buffer(data[:], maxSize)

	var id, event string
	var dataBuf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Bytes()
		line = bytes.TrimPrefix(line, bomBytes)

		if len(line) == 0 {
			if dataBuf.Len() > 0 {
				data := stripTrailingLF(dataBuf.Bytes())
				if err := f(id, event, data); err != nil {
					if err == ErrCloseEventStream {
						err = nil
					}
					return err
				}
				id = ""
				event = ""
				dataBuf.Reset()
			}

		} else if data, ok := bytes.CutPrefix(line, dataPrefixBytes); ok {
			dataBuf.Write(stripLeadingSpace(data))
			dataBuf.WriteByte('\n')

		} else if data, ok := bytes.CutPrefix(line, eventPrefixBytes); ok {
			event = string(stripLeadingSpace(data))

		} else if data, ok := bytes.CutPrefix(line, idPrefixBytes); ok {
			id = string(stripLeadingSpace(data))

		} else if data, ok := bytes.CutPrefix(line, retryPrefixBytes); ok {
			data = stripLeadingSpace(data)
			if asciiDigits.Match(data) {
				ms, err := strconv.ParseUint(string(stripLeadingSpace(data)), 10, 64)
				if err == nil && retryf != nil {
					retryf(ms)
				}
			}
		} // ignore unknown lines
	}
	// incomplete events are discarded
	return scanner.Err()
}

func stripLeadingSpace(data []byte) []byte {
	if len(data) > 0 && data[0] == ' ' {
		return data[1:]
	} else {
		return data
	}
}

func stripTrailingLF(data []byte) []byte {
	n := len(data)
	if n > 0 && data[n-1] == '\n' {
		return data[:n-1]
	} else {
		return data
	}
}

// scanLinesWithLF is like bufio.ScanLines, but accepts LF as a valid
// line separator in addition to CR and CR LF.
func scanLinesWithLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	i := bytes.IndexByte(data, '\n')
	j := bytes.IndexByte(data, '\r')

	if i >= 0 && (j < 0 || i < j) { // LF.
		// LF CR is not a valid separator, so no ambiguity here.
		return i + 1, data[0:i], nil

	} else if j >= 0 { // CR. could be part of CR LF.
		if j+1 < len(data) {
			if data[j+1] == '\n' {
				return j + 2, data[0:j], nil // found CR LF
			} else {
				return j + 1, data[0:j], nil // found LF
			}
		} else if atEOF {
			return j + 1, data[0:j], nil // found CR <EOF>
		} else {
			// CR is at end of buffer, can't tell if it is gonna be
			// followed by LF, so ask for more data.
			return 0, nil, nil
		}
	}

	if atEOF {
		return len(data), data, nil // final unterminated line
	}

	return 0, nil, nil
}
