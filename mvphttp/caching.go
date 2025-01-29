package mvphttp

import "net/http"

type CacheMode interface {
	EnumCacheHeaders(f func(key, value string))
}

type StandardCacheMode int

const (
	NoCacheHeaders StandardCacheMode = iota
	Uncached
	PrivateMutable
	PublicMutable
	PublicImmutable
)

func (cm StandardCacheMode) EnumCacheHeaders(f func(key, value string)) {
	switch cm {
	case NoCacheHeaders:
		break
	case Uncached:
		f("Expires", "Thu, 01 Jan 1970 00:00:00 UTC")
		f("Cache-Control", "no-cache, no-store, no-transform, must-revalidate, private, max-age=0")
		f("Pragma", "no-cache")
	case PrivateMutable:
		f("Cache-Control", "private, no-cache, max-age=0")
	case PublicMutable:
		f("Cache-Control", "public, no-cache, max-age=0")
	case PublicImmutable:
		f("Cache-Control", "public, max-age=31536000, immutable")
	default:
		panic("unknown StandardCacheMode")
	}
}

func ApplyCacheMode(w http.ResponseWriter, m CacheMode) {
	ApplyCacheModeToExistingHeader(w.Header(), m)
}

func ApplyCacheModeToExistingHeader(h http.Header, m CacheMode) {
	m.EnumCacheHeaders(func(key, value string) {
		h.Set(key, value)
	})
}

func ApplyCacheModeToHeader(h *http.Header, m CacheMode) {
	m.EnumCacheHeaders(func(key, value string) {
		if h == nil {
			*h = make(http.Header, 5)
		}
		h.Set(key, value)
	})
}

func HeaderWithCacheMode(m CacheMode) http.Header {
	h := make(http.Header, 5)
	ApplyCacheModeToExistingHeader(h, m)
	return h
}
