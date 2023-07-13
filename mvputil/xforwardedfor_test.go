package mvputil

import (
	"fmt"
	"net/http"
	"strings"
)

func ExampleIPForwarding() {
	forwardingFrom127 := &IPForwarding{
		ProxyIPs: MustParseCIDRs("127.0.0.0/8"),
	}

	try := func(desc string, r *http.Request) {
		fmt.Printf("\n== %s from %v via %v:\n", desc, r.RemoteAddr, strings.Join(r.Header.Values(XForwardedFor), " | "))
		fmt.Printf("IPForwardingFrom127 => ")
		fmt.Println(forwardingFrom127.FromRequest(r))
		fmt.Printf("IPForwardingFromLocalhost => ")
		fmt.Println(IPForwardingFromLocalhost.FromRequest(r))
		fmt.Printf("IPForwardingFromNoone => ")
		fmt.Println(IPForwardingFromNoone.FromRequest(r))
	}

	try("direct request", &http.Request{
		RemoteAddr: "127.0.0.1:23456",
		Host:       "localhost",
		Header:     http.Header{},
	})

	try("localhost-proxied request", &http.Request{
		RemoteAddr: "127.0.0.1:23456",
		Host:       "localhost",
		Header: http.Header{
			XForwardedFor:   {"1.2.3.4"},
			XForwardedHost:  {"example.com"},
			XForwardedProto: {"https"},
		},
	})

	try("remotely-proxied request", &http.Request{
		RemoteAddr: "11.22.33.44:23456",
		Host:       "localhost",
		Header: http.Header{
			XForwardedFor:   {"1.2.3.4"},
			XForwardedHost:  {"example.com"},
			XForwardedProto: {"https"},
		},
	})

	try("request with multiple proxies", &http.Request{
		RemoteAddr: "127.0.0.1:23456",
		Host:       "localhost",
		Header: http.Header{
			XForwardedFor:   {"1.2.3.4, 127.1.2.3"},
			XForwardedHost:  {"example.com"},
			XForwardedProto: {"https"},
		},
	})

	try("request with multiple proxies in several headers", &http.Request{
		RemoteAddr: "127.0.0.1:23456",
		Host:       "localhost",
		Header: http.Header{
			XForwardedFor:   {"1.2.3.4, 1.2.3.5", "1.2.3.6", "127.1.2.3"},
			XForwardedHost:  {"example.com"},
			XForwardedProto: {"https"},
		},
	})

	// Output:
	// == direct request from 127.0.0.1:23456 via :
	// IPForwardingFrom127 => 127.0.0.1 localhost false
	// IPForwardingFromLocalhost => 127.0.0.1 localhost false
	// IPForwardingFromNoone => 127.0.0.1 localhost false
	//
	// == localhost-proxied request from 127.0.0.1:23456 via 1.2.3.4:
	// IPForwardingFrom127 => 1.2.3.4 example.com true
	// IPForwardingFromLocalhost => 1.2.3.4 example.com true
	// IPForwardingFromNoone => 127.0.0.1 localhost false
	//
	// == remotely-proxied request from 11.22.33.44:23456 via 1.2.3.4:
	// IPForwardingFrom127 => 11.22.33.44 localhost false
	// IPForwardingFromLocalhost => 11.22.33.44 localhost false
	// IPForwardingFromNoone => 11.22.33.44 localhost false
	//
	// == request with multiple proxies from 127.0.0.1:23456 via 1.2.3.4, 127.1.2.3:
	// IPForwardingFrom127 => 1.2.3.4 example.com true
	// IPForwardingFromLocalhost => 127.1.2.3 example.com true
	// IPForwardingFromNoone => 127.0.0.1 localhost false
	//
	// == request with multiple proxies in several headers from 127.0.0.1:23456 via 1.2.3.4, 1.2.3.5 | 1.2.3.6 | 127.1.2.3:
	// IPForwardingFrom127 => 1.2.3.6 example.com true
	// IPForwardingFromLocalhost => 127.1.2.3 example.com true
	// IPForwardingFromNoone => 127.0.0.1 localhost false
}
