package mvp

import (
	"fmt"
	"time"

	"github.com/andreyvit/mvp/flogger"
	"golang.org/x/time/rate"
)

type RateLimitSettings struct {
	PerSec rate.Limit
	Burst  int
}

type RateLimitPreset string

const (
	RateLimitPresetNone            = RateLimitPreset("none")
	RateLimitPresetDefaultWebRead  = RateLimitPreset("web.r")
	RateLimitPresetDefaultWebWrite = RateLimitPreset("web.w")
	RateLimitPresetDefaultAPIRead  = RateLimitPreset("api.r")
	RateLimitPresetDefaultAPIWrite = RateLimitPreset("api.w")
	RateLimitPresetSpam            = RateLimitPreset("spam")
	RateLimitPresetAuthentication  = RateLimitPreset("auth")
	RateLimitPresetDangerous       = RateLimitPreset("danger")
)

type RateLimitGranularity string

const (
	RateLimitGranularityApp  = RateLimitGranularity("app")
	RateLimitGranularityIP   = RateLimitGranularity("ip")
	RateLimitGranularityUser = RateLimitGranularity("user")
	RateLimitGranularityKey  = RateLimitGranularity("key")
)

type RateLimiter struct {
	defaultLimiter *rate.Limiter
	Preset         RateLimitPreset
	Settings       RateLimitSettings
}

func (limiter *RateLimiter) Limiter(key string) *rate.Limiter {
	if key == "" {
		return limiter.defaultLimiter
	} else {
		panic("TODO: add support for per-key limiters")
	}
}

func initRateLimiting(app *App) {
	app.rateLimiters = make(map[RateLimitPreset]map[RateLimitGranularity]*RateLimiter)
	for preset, granSettings := range app.Settings.RateLimits {
		granLimiters := make(map[RateLimitGranularity]*RateLimiter)
		for gran, settings := range granSettings {
			granLimiters[gran] = &RateLimiter{
				defaultLimiter: rate.NewLimiter(settings.PerSec, settings.Burst),
				Preset:         preset,
				Settings:       settings,
			}
		}
		app.rateLimiters[preset] = granLimiters
	}
	app.rateLimiters[RateLimitPresetNone] = map[RateLimitGranularity]*RateLimiter{}
}

func (app *App) RateLimiters(preset RateLimitPreset) map[RateLimitGranularity]*RateLimiter {
	result := app.rateLimiters[preset]
	if result == nil {
		panic(fmt.Errorf("rate limiter preset not configured: %q", preset))
	}
	return result
}

func (app *App) enforceRateLimit(rc *RC) (any, error) {
	if app.Settings.DisableRateLimits {
		return nil, nil
	}
	preset := rc.RateLimitPreset
	limiters := app.RateLimiters(preset)

	var maxDelay time.Duration
	var maxGran RateLimitGranularity

	enforce := func(gran RateLimitGranularity, key string) error {
		if limiter := limiters[gran]; limiter != nil {
			now := rc.Now()
			rsrv := limiter.Limiter(key).ReserveN(now, 1)
			delay := rsrv.DelayFrom(now)
			if delay > app.Settings.MaxRateLimitRequestDelay.Value() {
				rsrv.CancelAt(now)
				flogger.Log(rc, "ratelimit: %s:%s hard rate limit exceeded (refusing to sleep for %d ms)", preset, gran, delay.Milliseconds())
				return ErrTooManyRequests
			}
			if delay > maxDelay {
				maxDelay, maxGran = delay, gran
			}
		}
		return nil
	}
	if err := enforce(RateLimitGranularityApp, ""); err != nil {
		return nil, err
	}
	// TODO: IP, etc -- when we add support for per-key limiters

	if maxDelay > 0 {
		flogger.Log(rc, "ratelimit: %s:%s soft rate limit exceeded (delay %d ms)", preset, maxGran, maxDelay.Milliseconds())
		time.Sleep(maxDelay)
	}
	return nil, nil
}
