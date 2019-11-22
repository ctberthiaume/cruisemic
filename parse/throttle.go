package parse

import (
	"time"
)

// Throttle enforces per-feed rate limits on Data structs. If a Data is seen
// within interval seconds of the last accepted Data of that feed type its
// Throttled field is set to true. Throttle also keeps track of the most recent
// non-rate limited timestamp for each feed.
type Throttle struct {
	recent   map[string]time.Time
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
	th.recent = make(map[string]time.Time)
	return th
}

// Limit marks Data as Throttled if the time since the last non-throttled Data
// for this feed is >= 0 and < Throttle.interval. Does nothing if the feed is
// an empty string, i.e. unparsed data. Data older than the last non-throttled
// Data or newer than or equal to (last non-throttled Data + interval) will
// not be throttled and will update the recent timestamp for this feed.
func (th Throttle) Limit(d *Data) {
	if d.Feed == "" {
		return
	}
	feedt := th.Recent(d.Feed)
	if feedt.IsZero() {
		// Never seen this feed before
		th.recent[d.Feed] = d.Time
	} else {
		diff := d.Time.Sub(feedt)

		switch {
		case diff >= 0 && diff < th.interval:
			d.Throttled = true
		default:
			// Either this is >= interval since last data or we've gone
			// backward in time, in which case update time and don't throttle.
			// Important not to throttle and to reset recent time in backward
			// case to protect against incorrect timestamps far in the future
			// causing all subsequent data to be throttled, e.g. instead of
			// 2019 for year we see 20192019. No new data would every make it
			// past throttling if we didn't reset time at the next out of order
			// 2019 data point.
			th.recent[d.Feed] = d.Time
		}
	}
}

// Recent returns the most recent non-rate limited time seen for a feed, or if
// feed is an empty string then the most recent time for all feeds. If the feed
// doesn't exist, returns a zero value time.
func (th Throttle) Recent(feed string) (t time.Time) {
	if feed == "" {
		for _, feedt := range th.recent {
			if t.Before(feedt) {
				t = feedt
			}
		}
	} else {
		t = th.recent[feed]
	}
	return t
}
