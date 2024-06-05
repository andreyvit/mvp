package postmark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Message struct {
	From          string `json:"From"`
	To            string `json:"To"`
	ReplyTo       string `json:"ReplyTo,omitempty"`
	Cc            string `json:"Cc,omitempty"`
	Bcc           string `json:"Bcc,omitempty"`
	Subject       string `json:"Subject"`
	Tag           string `json:"Tag,omitempty"`
	TextBody      string `json:"TextBody,omitempty"`
	HtmlBody      string `json:"HtmlBody,omitempty"`
	MessageStream string `json:"MessageStream"`
	TemplateId    int64  `json:"TemplateId,omitempty"`
	TemplateModel any    `json:"TemplateModel,omitempty"`
}

type Credentials struct {
	ServerAccessToken string
}

type Caller struct {
	HTTPClient *http.Client
	Credentials
}

func (c *Caller) Send(ctx context.Context, msg *Message) error {
	return c.call(ctx, "Send", msg)
}

func (c *Caller) call(ctx context.Context, callID string, input *Message) error {
	inputRaw := must(json.Marshal(input))
	r := must(http.NewRequestWithContext(ctx, "POST", "https://api.postmarkapp.com/email", bytes.NewReader(inputRaw)))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")
	if c.ServerAccessToken != "" {
		r.Header.Set("X-Postmark-Server-Token", c.ServerAccessToken)
	}

	log.Printf("postmark.%s: %s", callID, curl(r.Method, r.URL.String(), r.Header, inputRaw))

	resp, err := c.HTTPClient.Do(r)
	if err != nil {
		return &Error{
			CallID:    callID,
			IsNetwork: true,
			Cause:     err,
		}
	}
	defer resp.Body.Close()

	outputRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Error{
			CallID:    callID,
			IsNetwork: true,
			Cause:     err,
		}
	}

	var body response
	err = json.Unmarshal(outputRaw, &resp)
	if err != nil {
		return &Error{
			CallID:            callID,
			IsNetwork:         len(outputRaw) == 0 || outputRaw[0] != '{',
			StatusCode:        resp.StatusCode,
			Message:           "error unmashalling body",
			RawResponseBody:   outputRaw,
			PrintResponseBody: true,
			Cause:             err,
		}
	}

	if (resp.StatusCode >= 200 && resp.StatusCode <= 299) && body.ErrorCode == 0 {
		return nil
	} else {
		return &Error{
			CallID:            callID,
			IsNetwork:         len(outputRaw) == 0 || outputRaw[0] != '{',
			StatusCode:        resp.StatusCode,
			ErrorCode:         body.ErrorCode,
			Message:           body.Message,
			RawResponseBody:   outputRaw,
			PrintResponseBody: true,
			Cause:             err,
		}
	}
}

type response struct {
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func curl(method, path string, headers http.Header, body []byte) string {
	var buf strings.Builder
	buf.WriteString("curl")
	buf.WriteString(" -i")
	if method != "GET" {
		buf.WriteString(" -X")
		buf.WriteString(method)
	} else {
		body = nil
	}
	for k, vv := range headers {
		for _, v := range vv {
			buf.WriteString(" -H '")
			buf.WriteString(k)
			buf.WriteString(": ")
			buf.WriteString(v)
			buf.WriteString("'")
		}
	}
	if body != nil {
		buf.WriteString(" -d ")
		buf.WriteString(shellQuote(string(body)))
	}
	buf.WriteString(" '")
	buf.WriteString(path)
	buf.WriteString("'")
	return buf.String()
}

func shellQuote(source string) string {
	const specialChars = "\\'\"`${[|&;<>()*?! \t\n~"
	const specialInDouble = "$\\\"!"

	var buf strings.Builder
	buf.Grow(len(source) + 10)

	// pick quotation style, preferring single quotes
	if !strings.ContainsAny(source, specialChars) {
		buf.WriteString(source)
	} else if !strings.ContainsRune(source, '\'') {
		buf.WriteByte('\'')
		buf.WriteString(source)
		buf.WriteByte('\'')
	} else {
		buf.WriteByte('"')
		for {
			i := strings.IndexAny(source, specialInDouble)
			if i < 0 {
				break
			}
			buf.WriteString(source[:i])
			buf.WriteByte('\\')
			buf.WriteByte(source[i])
			source = source[i+1:]
		}
		buf.WriteString(source)
		buf.WriteByte('"')
	}
	return buf.String()
}

type Error struct {
	CallID            string
	IsNetwork         bool
	StatusCode        int
	ErrorCode         int
	Message           string
	RawResponseBody   []byte
	PrintResponseBody bool
	Cause             error
}

func (e *Error) Error() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "%s: HTTP %d", e.CallID, e.StatusCode)
	if e.IsNetwork {
		buf.WriteString("network: ")
	}
	if e.ErrorCode != 0 {
		buf.WriteString(": ")
		buf.WriteString(strconv.Itoa(e.ErrorCode))
	}
	if e.Message != "" {
		buf.WriteString(": ")
		buf.WriteString(e.Message)
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

func (e *Error) Unwrap() error {
	return e.Cause
}
