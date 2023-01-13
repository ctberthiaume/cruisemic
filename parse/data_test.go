package parse

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilledData(t *testing.T) {
	assert := assert.New(t)
	t0, _ := time.Parse(time.RFC3339, "2019-08-21T00:00:00.5Z")
	d := Data{Time: t0, Values: []string{"a", "b"}, Errors: []error{fmt.Errorf("e1"), fmt.Errorf("e2")}}
	assert.Equal("::2019-08-21T00:00:00.5Z,a,b::e1,e2", d.String(), "filled Data.String()")
	assert.Equal("2019-08-21T00:00:00.5Z,a,b", d.Line(","), "filled Data.Line(',')")
	assert.True(d.OK(), "filled Data.OK() == true")
}

func TestEmptyData(t *testing.T) {
	assert := assert.New(t)
	d := Data{}
	assert.Equal("::", d.String(), "empty Data.String()")
	assert.Equal("0001-01-01T00:00:00Z", d.Line(","), "empty Data.Line(',')")
	assert.False(d.OK(), "empty Data.OK() == false")
}

func TestThrottledData(t *testing.T) {
	assert := assert.New(t)
	t0, _ := time.Parse(time.RFC3339, "2019-08-21T00:00:00.5Z")
	d := Data{Time: t0, Values: []string{"a", "b"}, Errors: []error{fmt.Errorf("e1"), fmt.Errorf("e2")}, Throttled: true}
	assert.Equal("(T)::2019-08-21T00:00:00.5Z,a,b::e1,e2", d.String(), "throttled Data.String()")
	assert.Equal("2019-08-21T00:00:00.5Z,a,b", d.Line(","), "throttled Data.Line()")
	assert.False(d.OK(), "throttled Data.OK() == false")
}
