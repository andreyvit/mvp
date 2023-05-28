package mvp

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type Site struct {
	id string
}

func NewSite(id string) *Site {
	return &Site{id}
}

var (
	NoSite      = NewSite("none")
	DefaultSite = NewSite("default")
)

type DomainRouter struct {
	exact     map[string]any
	wildcards map[domainWildcard]any
	wildcard  any
}

func (router *DomainRouter) ValidDomains() []string {
	var result []string
	for e := range router.exact {
		result = append(result, e)
	}
	for w := range router.wildcards {
		result = append(result, w.String())
	}
	if router.wildcard != nil {
		result = append(result, "*")
	}
	sort.Strings(result)
	return result
}

func (router *DomainRouter) Handler(domain string, h http.Handler) {
	router.set(domain, h)
}

func (router *DomainRouter) HandleFunc(domain string, h func(http.ResponseWriter, *http.Request)) {
	router.Handler(domain, http.HandlerFunc(h))
}

func (router *DomainRouter) Site(domain string, site *Site) {
	router.set(domain, site)
}

func (router *DomainRouter) find(domain string) any {
	if d := router.exact[domain]; d != nil {
		return d
	}
	n := len(domain)
	for w, d := range router.wildcards {
		if n >= len(w.prefix)+len(w.suffix) && strings.HasPrefix(domain, w.prefix) && strings.HasSuffix(domain, w.suffix) {
			return d
		}
	}
	return router.wildcard
}

func (router *DomainRouter) set(domain string, v any) {
	if domain == "" {
		return
	}
	if domain == "*" {
		router.wildcard = v
	} else if prefix, suffix, found := strings.Cut(domain, "*"); found {
		if strings.Count(domain, "*") != 1 {
			panic(fmt.Errorf("domain cannot have more than one *: %q", domain))
		}
		w := domainWildcard{prefix, suffix}
		if router.wildcards == nil {
			router.wildcards = make(map[domainWildcard]any)
		}
		router.wildcards[w] = v
	} else {
		if router.exact == nil {
			router.exact = make(map[string]any)
		}
		router.exact[domain] = v
	}
}

type domainWildcard struct {
	prefix string
	suffix string
}

func (w domainWildcard) String() string {
	return w.prefix + "*" + w.suffix
}
