package mvp

import "time"

var (
	// Now is the start time of a request. For short requests, it should be the Now
	// value for all processing that happens inside. Can be mocked.
	Now = Value[time.Time]("now")

	// WallStart is an unmocked start time of the request for
	RealStartTime = Value[time.Time]("real_start_time")
)

func (app *App) Now() time.Time {
	return time.Now()
}
