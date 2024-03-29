// Package httpcall takes boilerplate out of typical non-streaming HTTP API calls.
package httpcall

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	DefaultMaxResponseLength int64         = 100 * 1024 * 1024 // 100 MB seems like a fine safety limit
	DefaultRetryDelay        time.Duration = 500 * time.Millisecond
)

var (
	ErrResponseTooLong = errors.New("response too long")
)

type (
	Request struct {
		Context context.Context
		CallID  string

		Method                 string
		BaseURL                string
		Path                   string
		QueryParams            url.Values
		Input                  any
		RawRequestBody         []byte
		RequestBodyContentType string
		Headers                http.Header
		BasicAuth              BasicAuth
		HTTPRequest            *http.Request

		OutputPtr         any
		MaxResponseLength int64

		HTTPClient     *http.Client
		MaxAttempts    int
		RetryDelay     time.Duration
		LastRetryDelay time.Duration

		ShouldStart        func(r *Request) error
		Started            func(r *Request)
		ValidateOutput     func() error
		ParseResponse      func(r *Request) error
		ParseErrorResponse func(r *Request)
		Failed             func(r *Request)
		Finished           func(r *Request)

		UserObject any
		UserData   map[string]any

		Attempts        int
		HTTPResponse    *http.Response
		RawResponseBody []byte
		Error           *Error
		Duration        time.Duration

		initDone bool
	}

	BasicAuth struct {
		Username string
		Password string
	}
)

func (r *Request) IsIdempotent() bool {
	r.Init()
	return r.HTTPRequest.Method == http.MethodGet || r.HTTPRequest.Method == http.MethodHead
}

func (r *Request) StatusCode() int {
	if r.HTTPResponse == nil {
		return 0
	}
	return r.HTTPResponse.StatusCode
}

func (r *Request) OnShouldStart(f func(r *Request) error) {
	if f == nil {
		return
	}
	if prev := r.ShouldStart; prev != nil {
		r.ShouldStart = func(r *Request) error {
			if err := prev(r); err != nil {
				return err
			}
			return f(r)
		}
	} else {
		r.ShouldStart = f
	}
}

func (r *Request) OnStarted(f func(r *Request)) {
	if f == nil {
		return
	}
	if prev := r.Started; prev != nil {
		r.Started = func(r *Request) {
			prev(r)
			f(r)
		}
	} else {
		r.Started = f
	}
}

func (r *Request) OnFailed(f func(r *Request)) {
	if f == nil {
		return
	}
	if prev := r.Failed; prev != nil {
		r.Failed = func(r *Request) {
			f(r)
			prev(r)
		}
	} else {
		r.Failed = f
	}
}

func (r *Request) OnFinished(f func(r *Request)) {
	if f == nil {
		return
	}
	if prev := r.Finished; prev != nil {
		r.Finished = func(r *Request) {
			f(r)
			prev(r)
		}
	} else {
		r.Finished = f
	}
}

func (r *Request) OnValidate(f func() error) {
	if f == nil {
		return
	}
	if prev := r.ValidateOutput; prev != nil {
		r.ValidateOutput = func() error {
			if err := prev(); err != nil {
				return err
			}
			return f()
		}
	} else {
		r.ValidateOutput = f
	}
}

func (r *Request) Init() {
	if r.initDone {
		return
	}
	r.initDone = true

	if r.Context == nil {
		r.Context = context.Background()
	}
	if r.HTTPRequest == nil {
		urlStr := buildURL(r.BaseURL, r.Path, r.QueryParams).String()

		if r.Method == "" {
			panic("Method must be specified (or HTTPRequest)")
		}
		if r.BaseURL == "" && r.Path == "" {
			panic("BaseURL and/or Path must be specified (or HTTPRequest)")
		}

		if r.Input != nil || r.RawRequestBody != nil {
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				panic("GET incompatible with body (Input / RawRequestBody)")
			}

			if r.RawRequestBody == nil {
				if v, ok := r.Input.(url.Values); ok {
					r.RawRequestBody = []byte(v.Encode())
					if r.RequestBodyContentType == "" {
						r.RequestBodyContentType = "application/x-www-form-urlencoded; charset=utf-8"
					}
				} else {
					r.RawRequestBody = must(json.Marshal(r.Input))
					if r.RequestBodyContentType == "" {
						r.RequestBodyContentType = "application/json"
					}
				}
			}

			r.HTTPRequest = must(http.NewRequestWithContext(r.Context, r.Method, urlStr, bytes.NewReader(r.RawRequestBody)))
			if r.RequestBodyContentType != "" {
				r.HTTPRequest.Header.Set("Content-Type", r.RequestBodyContentType)
			}
		} else {
			r.HTTPRequest = must(http.NewRequestWithContext(r.Context, r.Method, urlStr, nil))

			// We allow setting hr.RequestBodyContentType even when it's not necessary,
			// so that it's easy to provide a default value without checking the method.
		}
	} else {
		if r.Method == "" {
			r.Method = r.HTTPRequest.Method
		}
		if r.Path == "" {
			r.Path = r.HTTPRequest.URL.Path
		}
	}

	if r.BasicAuth != (BasicAuth{}) {
		r.HTTPRequest.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	for k, vv := range r.Headers {
		r.HTTPRequest.Header[k] = vv
	}
}

func (r *Request) Do() error {
	r.Init()

	r.Attempts = 0 // reset in case we're doing same request twice
	for {
		r.Attempts++

		if r.ShouldStart != nil {
			if err := r.ShouldStart(r); err != nil {
				return &Error{
					CallID: r.CallID,
					Cause:  err,
				}
			}
		}

		if r.Started != nil {
			r.Started(r)
		}

		start := time.Now()
		r.Error = r.doOnce()
		r.Duration = time.Since(start)

		if r.Error != nil && r.Failed != nil {
			r.Failed(r)
		}
		if r.Error == nil || !r.Error.IsRetriable || r.Attempts >= r.MaxAttempts {
			break
		}
		delay := r.Error.RetryDelay
		if delay == 0 {
			delay = r.RetryDelay
		}
		if delay == 0 {
			delay = DefaultRetryDelay
		}
		r.LastRetryDelay = delay
		sleep(r.Context, delay)
	}

	if r.Finished != nil {
		r.Finished(r)
	}

	if r.Error == nil {
		return nil // avoid returning non-nil error interface with nil *Error inside
	}
	return r.Error
}

func (r *Request) doOnce() *Error {
	client := r.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	maxLen := r.MaxResponseLength
	if maxLen == 0 {
		maxLen = DefaultMaxResponseLength
	}

	resp, err := client.Do(r.HTTPRequest)
	if err != nil {
		return &Error{
			CallID:      r.CallID,
			IsNetwork:   true,
			IsRetriable: true,
			Cause:       err,
		}
	}
	defer resp.Body.Close()
	r.HTTPResponse = resp

	if maxLen > 0 {
		r.RawResponseBody, err = io.ReadAll(&limitedReader{resp.Body, maxLen})
	} else {
		r.RawResponseBody, err = io.ReadAll(resp.Body)
	}
	if err != nil {
		isNetwork := !errors.Is(err, ErrResponseTooLong)
		return &Error{
			CallID:      r.CallID,
			IsNetwork:   isNetwork,
			IsRetriable: isNetwork,
			Cause:       err,
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		var err error
		var isNetwork bool
		if r.ParseResponse != nil {
			err = r.ParseResponse(r)
		} else if r.OutputPtr == nil {
			// ignore body
		} else if bytePtr, ok := r.OutputPtr.(*[]byte); ok {
			*bytePtr = r.RawResponseBody
		} else {
			err := json.Unmarshal(r.RawResponseBody, r.OutputPtr)
			if err != nil {
				isNetwork = (len(r.RawResponseBody) == 0 || (r.RawResponseBody[0] != '{' && r.RawResponseBody[0] != '['))
			}
		}
		if err != nil {
			return &Error{
				CallID:            r.CallID,
				IsNetwork:         isNetwork,
				IsRetriable:       isNetwork,
				StatusCode:        resp.StatusCode,
				Message:           "error parsing response",
				RawResponseBody:   r.RawResponseBody,
				PrintResponseBody: true,
				Cause:             err,
			}
		}
		if r.ValidateOutput != nil {
			err := r.ValidateOutput()
			if err != nil {
				return &Error{
					CallID:            r.CallID,
					IsNetwork:         false,
					IsRetriable:       false,
					StatusCode:        resp.StatusCode,
					RawResponseBody:   r.RawResponseBody,
					PrintResponseBody: true,
					Cause:             err,
				}
			}
		}
		return nil
	} else {
		r.Error = &Error{
			CallID:            r.CallID,
			StatusCode:        resp.StatusCode,
			IsRetriable:       (resp.StatusCode >= 500 && resp.StatusCode <= 599),
			RawResponseBody:   r.RawResponseBody,
			PrintResponseBody: true,
		}
		if r.ParseErrorResponse != nil {
			r.ParseErrorResponse(r)
			if r.Error == nil {
				return nil // avoid returning non-nil error interface with nil *Error inside
			}
		}
		return r.Error
	}
}

func buildURL(baseURL, path string, queryParams url.Values) *url.URL {
	if baseURL == "" {
		baseURL, path = path, ""
	}
	u := must(url.Parse(baseURL))
	u.Path = joinURLPath(u.Path, path)
	if len(queryParams) > 0 {
		u.RawQuery = queryParams.Encode()
	}
	return u
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func joinURLPath(base, url string) string {
	if base == "" {
		return url
	}
	if url == "" {
		return base
	}
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(url, "/")
}

func sleep(ctx context.Context, dur time.Duration) {
	ctxDone := ctx.Done()
	if ctxDone == nil {
		time.Sleep(dur)
	} else {
		timer := time.NewTimer(dur)
		select {
		case <-timer.C:
			break
		case <-ctx.Done():
			timer.Stop()
		}
	}
}

// Like io.LimitedReader, but returns ErrResponseTooLong instead of EOF.
type limitedReader struct {
	R io.Reader // underlying reader
	N int64     // max bytes remaining
}

func (l *limitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, ErrResponseTooLong
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}
