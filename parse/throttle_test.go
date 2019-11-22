package parse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestThrottleRecent(t *testing.T) {
	assert := assert.New(t)
	dur, err := time.ParseDuration("10s")
	if err != nil {
		panic(err)
	}
	th := NewThrottle(dur)
	assert.Equal(time.Time{}, th.Recent(""))
	assert.Equal(time.Time{}, th.Recent("unknownFeed"))

	t0 := time.Now()
	th.recent["feed1"] = t0
	assert.Equal(t0, th.Recent(""))
	assert.Equal(t0, th.Recent("feed1"))

	t1 := time.Now()
	th.recent["feed1"] = t1
	assert.Equal(t1, th.Recent(""))
	assert.Equal(t1, th.Recent("feed1"))

	t2 := time.Now()
	th.recent["feed2"] = t2
	assert.Equal(t2, th.Recent(""))
	assert.Equal(t1, th.Recent("feed1"))
	assert.Equal(t2, th.Recent("feed2"))
}

func TestThrottleLimitOneFeed(t *testing.T) {
	assert := assert.New(t)

	// Ten second rate limit per feed
	dur, err := time.ParseDuration("10s")
	if err != nil {
		panic(err)
	}
	th := NewThrottle(dur)

	t0 := time.Date(2017, 6, 17, 0, 30, 29, 365000000, time.UTC)
	d0 := Data{Feed: "feed1", Time: t0}
	th.Limit(&d0)
	assert.Equal(t0, th.Recent("feed1"), "first feed1 updates feed1 time")
	assert.False(d0.Throttled, "first feed1 Data is not marked as Throttled")

	t1 := time.Date(2017, 6, 17, 0, 30, 38, 365000000, time.UTC)
	d1 := Data{Feed: "feed1", Time: t1}
	th.Limit(&d1)
	assert.Equal(t0, th.Recent("feed1"), "throttled feed1 Data doesn't update feed1 time")
	assert.True(d1.Throttled, "throttled feed1 Data is marked as Throttled")

	t2 := time.Date(2017, 6, 17, 0, 30, 39, 365000000, time.UTC)
	d2 := Data{Feed: "feed1", Time: t2}
	th.Limit(&d2)
	assert.Equal(t2, th.Recent("feed1"), "unthrottled feed1 Data updates feed1 time")
	assert.False(d2.Throttled, "unthrottled feed1 Data is not marked as Throttled")

	t3 := time.Date(20172017, 6, 17, 0, 30, 39, 365000000, time.UTC)
	d3 := Data{Feed: "feed1", Time: t3}
	th.Limit(&d3)
	assert.Equal(t3, th.Recent("feed1"), "unthrottled far future feed1 Data updates feed1 time")
	assert.False(d3.Throttled, "unthrottled far future feed1 Data is not marked as Throttled")

	t4 := time.Date(2017, 6, 17, 0, 30, 42, 365000000, time.UTC)
	d4 := Data{Feed: "feed1", Time: t4}
	th.Limit(&d4)
	assert.Equal(t4, th.Recent("feed1"), "unthrottled past feed1 Data updates feed1 time")
	assert.False(d4.Throttled, "unthrottled past feed1 Data is not marked as Throttled")
}

func TestThrottleLimitTwoFeeds(t *testing.T) {
	assert := assert.New(t)

	// Ten second rate limit per feed
	dur, err := time.ParseDuration("10s")
	if err != nil {
		panic(err)
	}
	th := NewThrottle(dur)

	t0 := time.Date(2017, 6, 17, 0, 30, 29, 365000000, time.UTC)
	d0 := Data{Feed: "feed1", Time: t0}
	th.Limit(&d0)
	assert.Equal(t0, th.Recent("feed1"), "first feed1 updates feed1 time")
	assert.False(d0.Throttled, "first feed1 Data is not marked as Throttled")

	t1 := time.Date(2017, 6, 17, 0, 30, 32, 365000000, time.UTC)
	d1 := Data{Feed: "feed2", Time: t1}
	th.Limit(&d1)
	assert.Equal(t1, th.Recent("feed2"), "unthrottled feed2 Data updates feed2 time")
	assert.Equal(t0, th.Recent("feed1"), "unthrottled feed2 Data doesn't update feed1 time")
	assert.False(d1.Throttled, "unthrottled feed2 Data is not marked as Throttled")

	t2 := time.Date(2017, 6, 17, 0, 30, 40, 365000000, time.UTC)
	d2 := Data{Feed: "feed2", Time: t2}
	th.Limit(&d2)
	assert.Equal(t1, th.Recent("feed2"), "throttled feed2 Data doesn't update feed2 time")
	assert.Equal(t0, th.Recent("feed1"), "throttled feed2 Data doesn't update feed1 time")
	assert.True(d2.Throttled, "throttled feed2 Data is marked as Throttled")

	t3 := time.Date(2017, 6, 17, 0, 30, 40, 365000000, time.UTC)
	d3 := Data{Feed: "feed1", Time: t3}
	th.Limit(&d3)
	assert.Equal(t3, th.Recent("feed1"), "unthrottled feed1 Data updates feed1 time")
	assert.Equal(t1, th.Recent("feed2"), "unthrottled feed1 Data doesn't update feed2 time")
	assert.False(d3.Throttled, "unthrottled feed1 Data is not marked as Throttled")
}

func TestThrottleLimitInterval0(t *testing.T) {
	assert := assert.New(t)

	th := NewThrottle(0)

	t0 := time.Date(2017, 6, 17, 0, 30, 29, 365000000, time.UTC)
	d0 := Data{Feed: "feed1", Time: t0}
	th.Limit(&d0)
	assert.Equal(t0, th.Recent("feed1"), "first feed1 updates feed1 time")
	assert.False(d0.Throttled, "first feed1 Data is not marked as Throttled")

	t1 := time.Date(2017, 6, 17, 0, 30, 33, 365000000, time.UTC)
	d1 := Data{Feed: "feed1", Time: t1}
	th.Limit(&d1)
	assert.Equal(t1, th.Recent("feed1"), "unthrottled feed1 Data updates feed1 time")
	assert.False(d1.Throttled, "unthrottled feed1 Data is not marked as Throttled")

	t2 := time.Date(2017, 6, 17, 0, 20, 33, 365000000, time.UTC)
	d2 := Data{Feed: "feed1", Time: t2}
	th.Limit(&d2)
	assert.Equal(t2, th.Recent("feed1"), "unthrottled past feed1 Data updates feed1 time")
	assert.False(d2.Throttled, "unthrottled feed1 Data is not marked as Throttled")
}

func TestThrottleBadDuration(t *testing.T) {
	assert := assert.New(t)
	dur, err := time.ParseDuration("-2s")
	if err != nil {
		panic(err)
	}
	th := NewThrottle(dur)
	assert.Equal(time.Duration(0), th.interval, "Duration -2s converts to 0s")
}

func TestGlobalRecent(t *testing.T) {
	assert := assert.New(t)
	th := NewThrottle(0)
	// Add some data points
	t0 := time.Date(2017, 6, 17, 0, 30, 55, 365000000, time.UTC)
	th.Limit(&Data{Feed: "feed1", Time: t0})
	t1 := time.Date(2017, 6, 17, 0, 30, 50, 365000000, time.UTC)
	th.Limit(&Data{Feed: "feed2", Time: t1})

	// Check before and after reset
	assert.Equal(t0, th.Recent(""), "global recent is feed1 recent time")
	t2 := time.Date(2017, 6, 17, 0, 31, 00, 365000000, time.UTC)
	th.Limit(&Data{Feed: "feed2", Time: t2})
	assert.Equal(t2, th.Recent(""), "global recent is feed2 recent time")
}

func TestEmptyThrottle(t *testing.T) {
	assert := assert.New(t)
	th := NewThrottle(0)
	assert.Equal(time.Time{}, th.Recent(""), "no recent global time")
}
