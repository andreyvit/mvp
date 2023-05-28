package backoff

import (
	"fmt"
	"time"
)

func ExampleBackoff() {
	b := Backoff{
		ImmediateRetries: 1,

		FixedDelayRetries: 2,
		FixedDelay:        time.Second,

		BackoffRetries:  10,
		MaxBackoffDelay: time.Minute,
	}

	for a := 1; a <= 15; a++ {
		s := b.DelayAfter(a).Seconds()
		fmt.Printf("after %2d failures delay is %.0f seconds\n", a, s)
	}
	// Output: after  1 failures delay is 0 seconds
	// after  2 failures delay is 1 seconds
	// after  3 failures delay is 1 seconds
	// after  4 failures delay is 2 seconds
	// after  5 failures delay is 4 seconds
	// after  6 failures delay is 8 seconds
	// after  7 failures delay is 16 seconds
	// after  8 failures delay is 32 seconds
	// after  9 failures delay is 60 seconds
	// after 10 failures delay is 60 seconds
	// after 11 failures delay is 60 seconds
	// after 12 failures delay is 60 seconds
	// after 13 failures delay is 60 seconds
	// after 14 failures delay is 1000000000 seconds
	// after 15 failures delay is 1000000000 seconds
}
