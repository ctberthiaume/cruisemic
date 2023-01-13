package parse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestThrottleLimit(t *testing.T) {
	assert := assert.New(t)

	// Ten second rate limit per feed
	dur, err := time.ParseDuration("10s")
	if err != nil {
		panic(err)
	}
	th := NewThrottle(dur)

	t0 := time.Date(2017, 6, 17, 0, 30, 29, 365000000, time.UTC)
	d0 := Data{Time: t0}
	th.Limit(&d0)
	assert.Equal(t0, th.recent, "first Data updates recent time")
	assert.False(d0.Throttled, "first Data is not marked as Throttled")

	t1 := time.Date(2017, 6, 17, 0, 30, 38, 365000000, time.UTC)
	d1 := Data{Time: t1}
	th.Limit(&d1)
	assert.Equal(t0, th.recent, "throttled Data doesn't update recent time")
	assert.True(d1.Throttled, "throttled Data is marked as Throttled")

	t2 := time.Date(2017, 6, 17, 0, 30, 39, 365000000, time.UTC)
	d2 := Data{Time: t2}
	th.Limit(&d2)
	assert.Equal(t2, th.recent, "unthrottled Data updates recent time")
	assert.False(d2.Throttled, "unthrottled Data is not marked as Throttled")

	t3 := time.Date(20172017, 6, 17, 0, 30, 39, 365000000, time.UTC)
	d3 := Data{Time: t3}
	th.Limit(&d3)
	assert.Equal(t3, th.recent, "unthrottled far future Data updates recent time")
	assert.False(d3.Throttled, "unthrottled far future Data is not marked as Throttled")

	t4 := time.Date(2017, 6, 17, 0, 30, 42, 365000000, time.UTC)
	d4 := Data{Time: t4}
	th.Limit(&d4)
	assert.Equal(t4, th.recent, "unthrottled past Data updates recent time")
	assert.False(d4.Throttled, "unthrottled past Data is not marked as Throttled")
}

func TestThrottleLimitInterval0(t *testing.T) {
	assert := assert.New(t)

	th := NewThrottle(0)

	t0 := time.Date(2017, 6, 17, 0, 30, 29, 365000000, time.UTC)
	d0 := Data{Time: t0}
	th.Limit(&d0)
	assert.Equal(t0, th.recent, "first Data updates recent time")
	assert.False(d0.Throttled, "first Data is not marked as Throttled")

	t1 := time.Date(2017, 6, 17, 0, 30, 33, 365000000, time.UTC)
	d1 := Data{Time: t1}
	th.Limit(&d1)
	assert.Equal(t1, th.recent, "unthrottled Data updates recent time")
	assert.False(d1.Throttled, "unthrottled Data is not marked as Throttled")

	t2 := time.Date(2017, 6, 17, 0, 20, 33, 365000000, time.UTC)
	d2 := Data{Time: t2}
	th.Limit(&d2)
	assert.Equal(t2, th.recent, "unthrottled past Data updates recent time")
	assert.False(d2.Throttled, "unthrottled Data is not marked as Throttled")
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
