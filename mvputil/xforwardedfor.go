package mvputil

import (
	"net"
	"net/http"
	"strings"
)

type IPForwarding struct {
	ProxyIPs []*net.IPNet
}

const (
	XForwardedFor   = "X-Forwarded-For"
	XForwardedHost  = "X-Forwarded-Host"
	XForwardedProto = "X-Forwarded-Proto"
)

var (
	LoopbackIPv4CIRD = MustParseCIDR("127.0.0.1/32")
	LoopbackIPv6CIRD = MustParseCIDR("::1/128")
	AnyIPv4CIRD      = MustParseCIDR("0.0.0.0/0")
	AnyIPv6CIRD      = MustParseCIDR("::0/0")

	IPForwardingFromNoone     = &IPForwarding{}
	IPForwardingFromAnyone    = &IPForwarding{[]*net.IPNet{LoopbackIPv4CIRD, LoopbackIPv6CIRD}}
	IPForwardingFromLocalhost = &IPForwarding{[]*net.IPNet{LoopbackIPv4CIRD, LoopbackIPv6CIRD}}
)

func MustParseCIDRs(s string) []*net.IPNet {
	items := strings.Fields(s)
	result := make([]*net.IPNet, 0, len(items))
	for _, item := range items {
		result = append(result, MustParseCIDR(item))
	}
	return result
}

func MustParseCIDR(s string) *net.IPNet {
	_, cidr, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return cidr
}

// TrimPort removes :port part (if any) from the given string and returns just the hostname.
func TrimPort(hostport string) string {
	h, _, err := net.SplitHostPort(hostport)
	if h == "" || err != nil {
		return hostport
	}
	return h
}

func (conf *IPForwarding) IsWhitelistedProxy(ip net.IP) bool {
	for _, cidr := range conf.ProxyIPs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func (conf *IPForwarding) FromRequest(r *http.Request) (ip net.IP, host string, isTLS bool) {
	ip = net.ParseIP(TrimPort(r.RemoteAddr))
	host = r.Host
	isTLS = (r.TLS != nil)

	if ip != nil && conf.IsWhitelistedProxy(ip) {
		if xfp := r.Header.Get(XForwardedProto); xfp != "" {
			isTLS = (xfp == "https")
		}
		if xfh := r.Header.Get(XForwardedHost); xfh != "" {
			host = xfh
		}

		xff := strings.Join(r.Header.Values(XForwardedFor), ",")
		comps := strings.FieldsFunc(xff, func(r rune) bool { return r == ',' })
		for i := len(comps) - 1; i >= 0; i-- {
			s := strings.TrimSpace(comps[i])
			if s == "" {
				continue
			}
			ip = net.ParseIP(s)
			if ip == nil {
				break
			}
			if i > 0 && !conf.IsWhitelistedProxy(ip) {
				break
			}
		}
	}

	return
}
