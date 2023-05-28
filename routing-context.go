package mvp

type ErrHandler = func(rc *RC, err error) (any, error)

type Router interface {
	Use(middleware any)
	UseIn(slot string, middleware any)
	OnError(f ErrHandler)
	RateLimit(preset RateLimitPreset)
}

type routingContext struct {
	middleware      middlewareSlotList
	errHandlers     []func(rc *RC, err error) (any, error)
	rateLimitPreset RateLimitPreset
}

func (c *routingContext) clone() routingContext {
	return routingContext{
		middleware:      c.middleware.Clone(),
		errHandlers:     append([]ErrHandler(nil), c.errHandlers...),
		rateLimitPreset: c.rateLimitPreset,
	}
}

// Use adds the given middleware to all future routes.
func (c *routingContext) Use(middleware any) {
	c.middleware.Add("", adaptMiddleware(middleware))
}

// UseIn adds the given middleware under the given slot. If the slot is already
// occupied, overrides the middleware in the given slot (so that you, for example,
// can define a default authentication or CORS middleware, but override it
// for certain groups or routes). Use nil as middleware to override with nothing,
// or to reserve a slot position without assigning middlware yet.
func (c *routingContext) UseIn(slot string, middleware any) {
	c.middleware.Add(slot, adaptMiddleware(middleware))
}
func (c *routingContext) OnError(f ErrHandler) {
	c.errHandlers = append(c.errHandlers, f)
}
func (c *routingContext) RateLimit(preset RateLimitPreset) {
	c.rateLimitPreset = preset
}
