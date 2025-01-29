package cors

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestCORSWildcard(t *testing.T) {
	c := &CORS{
		RequestHeaders: "Content-Type, User-Agent, Authorization",
	}

	tests := []struct {
		method    string
		origin    string
		preflight string
		request   string
	}{
		{
			method:    "",
			origin:    "",
			preflight: "400 missing Origin HTTP header | Vary: Origin | Vary: Access-Control-Request-Method",
			request:   "Access-Control-Allow-Origin: *",
		},
		{
			method:    "",
			origin:    "xxx",
			preflight: "400 missing Access-Control-Request-Method HTTP header | Vary: Origin | Vary: Access-Control-Request-Method",
			request:   "Access-Control-Allow-Origin: *",
		},
		{
			method:    "PUT",
			origin:    "xxx",
			preflight: "200 OK | Access-Control-Allow-Headers: Content-Type, User-Agent, Authorization | Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH | Access-Control-Allow-Origin: * | Vary: Origin | Vary: Access-Control-Request-Method",
			request:   "Access-Control-Allow-Origin: *",
		},
	}

	for _, test := range tests {
		actual := preflightCORS(c, test.origin, test.method)
		if actual != test.preflight {
			t.Errorf("** Preflight(%s %q) = %s, wanted %s", test.method, test.origin, actual, test.preflight)
		} else {
			t.Logf("✓ Preflight(%s %q) = %s", test.method, test.origin, actual)
		}

		actual = requestCORS(c, test.origin)
		if actual != test.request {
			t.Errorf("** Request(%s %q) = %s, wanted %s", test.method, test.origin, actual, test.request)
		} else {
			t.Logf("✓ Request(%s %q) = %s", test.method, test.origin, actual)
		}
	}
}

func TestCORSSpecific(t *testing.T) {
	c := &CORS{
		Origins:        []string{"https://example.com"},
		RequestHeaders: "Content-Type, User-Agent, Authorization",
	}

	tests := []struct {
		method    string
		origin    string
		preflight string
		request   string
	}{
		{
			method:    "",
			origin:    "",
			preflight: "400 missing Origin HTTP header | Vary: Origin | Vary: Access-Control-Request-Method",
			request:   "",
		},
		{
			method:    "",
			origin:    "xxx",
			preflight: "400 missing Access-Control-Request-Method HTTP header | Vary: Origin | Vary: Access-Control-Request-Method",
			request:   "",
		},
		{
			method:    "PUT",
			origin:    "xxx",
			preflight: "403 forbidden origin | Vary: Origin | Vary: Access-Control-Request-Method",
			request:   "",
		},
		{
			method:    "PUT",
			origin:    "https://example.com",
			preflight: "200 OK | Access-Control-Allow-Headers: Content-Type, User-Agent, Authorization | Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH | Access-Control-Allow-Origin: https://example.com | Vary: Origin | Vary: Access-Control-Request-Method",
			request:   "Access-Control-Allow-Origin: https://example.com | Vary: Origin",
		},
	}

	for _, test := range tests {
		actual := preflightCORS(c, test.origin, test.method)
		if actual != test.preflight {
			t.Errorf("** Preflight(%s %q) = %s, wanted %s", test.method, test.origin, actual, test.preflight)
		} else {
			t.Logf("✓ Preflight(%s %q) = %s", test.method, test.origin, actual)
		}

		actual = requestCORS(c, test.origin)
		if actual != test.request {
			t.Errorf("** Request(%s %q) = %s, wanted %s", test.method, test.origin, actual, test.request)
		} else {
			t.Logf("✓ Request(%s %q) = %s", test.method, test.origin, actual)
		}
	}
}

func requestCORS(c *CORS, origin string) string {
	hdr := http.Header{}
	c.AddCORSHeadersOrigin(origin, hdr)

	var hdrBuf strings.Builder
	hdr.Write(&hdrBuf)
	hdrStr := hdrBuf.String()

	return strings.ReplaceAll(strings.TrimSpace(hdrStr), "\r\n", " | ")
}

func preflightCORS(c *CORS, origin, method string) string {
	hdr := http.Header{}
	msg, status := c.computePreflightResponse(origin, method, hdr)
	if status == 200 {
		msg = "OK"
	}
	intro := fmt.Sprintf("%d %s", status, msg)

	var hdrBuf strings.Builder
	hdr.Write(&hdrBuf)
	hdrStr := hdrBuf.String()

	if hdrStr == "" {
		return intro
	} else {
		return intro + " | " + strings.ReplaceAll(strings.TrimSpace(hdrStr), "\r\n", " | ")
	}
}
