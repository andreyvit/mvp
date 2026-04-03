package postmark

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/andreyvit/mvp/httpcall"
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
	Credentials
	ConfigureHTTPRequest func(ctx context.Context, r *httpcall.Request)
}

func (c *Caller) Send(ctx context.Context, msg *Message) error {
	return c.call(ctx, "Send", msg)
}

func (c *Caller) call(ctx context.Context, callID string, input *Message) error {
	inputRaw := must(json.Marshal(input))

	var body response
	r := &httpcall.Request{
		Context:                ctx,
		CallID:                 callID,
		Method:                 "POST",
		Path:                   "https://api.postmarkapp.com/email",
		RawRequestBody:         inputRaw,
		RequestBodyContentType: "application/json",
		OutputPtr:              &body,
		MaxAttempts:            1,
		ParseErrorResponse:     parseErrorResponse,
		ValidateOutput: func() error {
			if body.ErrorCode != 0 {
				return &Error{
					CallID:    callID,
					ErrorCode: body.ErrorCode,
					Message:   body.Message,
				}
			}
			return nil
		},
	}
	r.SetHeader("Accept", "application/json")
	if c.ServerAccessToken != "" {
		r.SetHeader("X-Postmark-Server-Token", c.ServerAccessToken)
	}

	if c.ConfigureHTTPRequest != nil {
		c.ConfigureHTTPRequest(ctx, r)
	}
	return r.Do()
}

func parseErrorResponse(r *httpcall.Request) {
	var body response
	_ = json.Unmarshal(r.RawResponseBody, &body)
	if body.Message != "" {
		r.Error.Message = body.Message
	}
	if body.ErrorCode != 0 {
		r.Error.Type = strconv.Itoa(body.ErrorCode)
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
