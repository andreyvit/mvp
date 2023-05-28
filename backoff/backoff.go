package backoff

import (
	"fmt"
	"time"
)

const InfiniteDelay time.Duration = 1000000000 * time.Second

type Backoff struct {
	ImmediateRetries int

	FixedDelayRetries int
	FixedDelay        time.Duration

	BackoffRetries  int
	MinBackoffDelay time.Duration
	MaxBackoffDelay time.Duration
	BackoffFactor   float64
}

func (b Backoff) String() string {
	return fmt.Sprintf("%d + %d/%v + %d", b.ImmediateRetries, b.FixedDelayRetries, b.FixedDelay, b.BackoffRetries)
}

var GoodBackoff = Backoff{
	ImmediateRetries: 1,

	FixedDelayRetries: 2,
	FixedDelay:        time.Second,

	BackoffRetries:  50,
	MaxBackoffDelay: time.Hour,
}

var NoBackoff = Backoff{}

func (b Backoff) DelayAfter(failedAttempts int) time.Duration {
	if failedAttempts == 0 {
		panic("failed attempts cannot be zero")
	}
	if failedAttempts < 0 {
		panic("failed attempts cannot be negative")
	}
	retries := failedAttempts - 1

	if retries < b.ImmediateRetries {
		return 0
	}
	retries -= b.ImmediateRetries

	if retries < b.FixedDelayRetries {
		return b.FixedDelay
	}
	retries -= b.FixedDelayRetries

	if retries >= b.BackoffRetries {
		return InfiniteDelay
	}

	factor := b.BackoffFactor
	if factor == 0 {
		factor = 2.0
	}
	delay := b.MinBackoffDelay
	if delay == 0 {
		delay = time.Duration(factor*float64(b.FixedDelay) + 0.5)
	}
	for i := 0; i < retries && delay <= b.MaxBackoffDelay; i++ {
		delay = time.Duration(factor*float64(delay) + 0.5)
	}
	if delay > b.MaxBackoffDelay {
		delay = b.MaxBackoffDelay
	}
	return delay
}
