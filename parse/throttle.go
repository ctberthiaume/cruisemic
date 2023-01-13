package parse

import (
	"time"
)

// Throttle enforces rate limits on Data structs. If a Data is seen
// within interval seconds of the last unthrottled Data then its Throttled field
// is set to true.
type Throttle struct {
	recent   time.Time
	interval time.Duration
}

// NewThrottle creates a new Throttle struct. Use interval of 0s to turn off
// throttling. intervals < 0s will be set to 0s.
func NewThrottle(interval time.Duration) (th Throttle) {
	zero, err := time.ParseDuration("0s")
	if err != nil {
		panic(err)
	}
	if interval < zero {
		interval = zero
	}
	th.interval = interval
	return th
}

// Limit marks Data as Throttled if the time since the last non-throttled Data
// is >= 0 and < Throttle.interval. Data older than the last non-throttled
// Data or newer than or equal to (last non-throttled Data + interval) will
// not be throttled and will update the recent timestamp. Data with zero time
// will be ignored.
func (th *Throttle) Limit(d *Data) {
	if d.Time.IsZero() {
		return
	}
	if th.recent.IsZero() {
		th.recent = d.Time
	} else {
		diff := d.Time.Sub(th.recent)

		switch {
		case diff >= 0 && diff < th.interval:
			d.Throttled = true
		default:
			// Either this is >= interval since last data or we've gone
			// backward in time, in which case update time and don't throttle.
			// Important not to throttle and to reset recent time in backward
			// case to protect against incorrect timestamps far in the future
			// causing all subsequent data to be throttled, e.g. instead of
			// 2019 for year we see 20192019. No new data would ever make it
			// past throttling if we didn't reset time at the next out of order
			// 2019 data point.
			th.recent = d.Time
		}
	}
}
