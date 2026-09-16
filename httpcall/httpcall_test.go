package httpcall_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/andreyvit/mvp/httpcall"
)

func TestRequest_force_non_idempotent_read_only_policy(t *testing.T) {
	for _, tc := range []struct {
		name    string
		method  string
		force   bool
		blocked bool
	}{
		{name: "get", method: http.MethodGet},
		{name: "head", method: http.MethodHead},
		{name: "post", method: http.MethodPost, blocked: true},
		{name: "forced_get", method: http.MethodGet, force: true, blocked: true},
		{name: "forced_head", method: http.MethodHead, force: true, blocked: true},
		{name: "forced_post", method: http.MethodPost, force: true, blocked: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				readOnly    = true
				errReadOnly = errors.New("read-only")

				methods []string
				req     = &httpcall.Request{
					Method:             tc.method,
					BaseURL:            "https://example.test",
					ForceNonIdempotent: tc.force,
					HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
						methods = append(methods, r.Method)
						return &http.Response{
							StatusCode: http.StatusOK,
							Header:     make(http.Header),
							Body:       io.NopCloser(strings.NewReader("{}")),
							Request:    r,
						}, nil
					})},
				}
			)
			req.OnShouldStart(func(r *httpcall.Request) error {
				if readOnly && !r.IsIdempotent() {
					return errReadOnly
				}
				return nil
			})

			requests := []*httpcall.Request{req, req.Clone()}
			for _, r := range requests {
				err := r.Do()
				if tc.blocked {
					if !errors.Is(err, errReadOnly) {
						t.Fatalf("Do() = %v, want read-only", err)
					}
				} else if err != nil {
					t.Fatal(err)
				}
			}

			wantCalls := len(requests)
			if tc.blocked {
				wantCalls = 0
			}

			if len(methods) != wantCalls {
				t.Fatalf("requests sent = %d, want %d", len(methods), wantCalls)
			}

			readOnly = false
			for _, r := range requests {
				if err := r.Do(); err != nil {
					t.Fatal(err)
				}
			}

			if len(methods) != wantCalls+len(requests) {
				t.Fatalf("requests sent = %d, want %d", len(methods), wantCalls+len(requests))
			}

			for _, method := range methods {
				if method != tc.method {
					t.Errorf("HTTP method = %s, want %s", method, tc.method)
				}
			}
		})
	}
}

func TestRequest_force_non_idempotent_preserves_retries(t *testing.T) {
	var (
		calls int
		req   = &httpcall.Request{
			Method:             http.MethodGet,
			BaseURL:            "https://example.test",
			ForceNonIdempotent: true,
			MaxAttempts:        2,
			RetryDelay:         time.Nanosecond,
			HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != http.MethodGet {
					t.Errorf("HTTP method = %s, want GET", r.Method)
				}

				if calls == 1 {
					return nil, errors.New("temporary network failure")
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader("{}")),
					Request:    r,
				}, nil
			})},
		}
	)
	if err := req.Do(); err != nil {
		t.Fatal(err)
	}

	if calls != 2 {
		t.Errorf("requests sent = %d, want 2", calls)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
