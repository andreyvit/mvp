package longtermcache

import (
	"sync"
)

// TODO: protect from  memory-based DoS attacks.
// Attack: spam the app with requests to non-existent entries.
// Should discard those entries quickly as long as nobody is accessing them.
// But this is a bit tricky to get right.

type (
	Cache[K comparable, T any] struct {
		Make        func(K) *T
		entries     map[K]*entry[T]
		entriesLock sync.Mutex
	}

	Policy int

	entry[T any] struct {
		value    *T
		computed bool

		makingLock sync.RWMutex // obtain before valueLock
		valueLock  sync.RWMutex
	}
)

const (
	Available Policy = iota // Available accepts a stale value if available
	Fresh                   // Fresh insists on waiting for the latest value if update is running
	UpdateNow               // UpdateNow initiates an update
)

var _policyStrings = [...]string{
	Available: "available",
	Fresh:     "fresh",
	UpdateNow: "now",
}

func (v Policy) String() string {
	return _policyStrings[v]
}

func (c *Cache[K, T]) Lookup(key K, pol Policy) *T {
	e := c.lookupEntry(key)

	switch pol {
	case Available:
		if value, ok := e.load(); ok {
			return value
		}
		fallthrough
	case Fresh:
		e.makingLock.Lock()
		defer e.makingLock.Unlock()
		if value, ok := e.load(); ok {
			return value
		}
		return c.updateHoldingMakingLock(key, e)

	case UpdateNow:
		e.makingLock.Lock()
		defer e.makingLock.Unlock()
		return c.updateHoldingMakingLock(key, e)

	default:
		panic("invalid policy")
	}
}

func (c *Cache[K, T]) Discard(key K) {
	c.lookupEntry(key).discard()
}

func (c *Cache[K, T]) updateHoldingMakingLock(key K, e *entry[T]) *T {
	if c.Make == nil {
		panic("make func not set")
	}
	value := c.Make(key)
	e.store(value)
	return value
}

func (c *Cache[K, T]) lookupEntry(key K) *entry[T] {
	// NOTE: this can easily be made faster by using sync.Map or similar approach
	c.entriesLock.Lock()
	defer c.entriesLock.Unlock()
	e := c.entries[key]
	if e == nil {
		if c.entries == nil {
			c.entries = make(map[K]*entry[T])
		}
		e = &entry[T]{}
		c.entries[key] = e
	}
	return e
}

func (e *entry[T]) load() (*T, bool) {
	e.valueLock.Lock()
	defer e.valueLock.Unlock()
	return e.value, e.computed
}
func (e *entry[T]) store(value *T) {
	e.valueLock.Lock()
	defer e.valueLock.Unlock()
	e.value, e.computed = value, true
}
func (e *entry[T]) discard() {
	e.valueLock.Lock()
	defer e.valueLock.Unlock()
	e.value, e.computed = nil, false
}
