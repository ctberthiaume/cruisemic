package parse

import (
	"testing"
	"time"

	"github.com/ctberthiaume/tsdata"
	"github.com/stretchr/testify/assert"
)

func TestFilledData(t *testing.T) {
	assert := assert.New(t)
	t0, _ := time.Parse(time.RFC3339, "2019-08-21T00:00:00.5Z")
	d := Data{Feed: "feedType", Time: t0, Values: []string{"a", "b"}}
	assert.Equal("feedType: 2019-08-21T00:00:00.5Z,a,b", d.String(), "filled Data.String()")
	assert.Equal("2019-08-21T00:00:00.5Z,a,b", d.Line(","), "filled Data.Line(',')")
	assert.True(d.OK(), "filled Data.OK() == true")
}

func TestEmptyData(t *testing.T) {
	assert := assert.New(t)
	d := Data{}
	assert.Equal(": ", d.String(), "empty Data.String()")
	assert.Equal("0001-01-01T00:00:00Z", d.Line(","), "empty Data.Line(',')")
	assert.False(d.OK(), "empty Data.OK() == false")

	d = Data{Feed: "feedType"}
	assert.Equal("feedType: ", d.String(), "empty with Feed Data.String()")
	assert.Equal("0001-01-01T00:00:00Z", d.Line(","), "empty with Feed Data.Line(',')")
	assert.False(d.OK(), "empty with Feed Data.OK() == false")
}

func TestThrottledData(t *testing.T) {
	assert := assert.New(t)
	t0, _ := time.Parse(time.RFC3339, "2019-08-21T00:00:00.5Z")
	d := Data{Feed: "feedType", Time: t0, Values: []string{"a", "b"}, Throttled: true}
	assert.Equal("feedType(T): 2019-08-21T00:00:00.5Z,a,b", d.String(), "throttled Data.String()")
	assert.Equal("2019-08-21T00:00:00.5Z,a,b", d.Line(","), "throttled Data.Line()")
	assert.False(d.OK(), "throttled Data.OK() == false")
}

var wantHeader = `testfeed
testproject
Test comment
time of exp	a good value
time	float
NA	parsecs
time	testcol`

func TestDataDescHeaders(t *testing.T) {
	assert := assert.New(t)
	d := NewFeedCollection()
	d.Feeds["testfeed"] = tsdata.Tsdata{
		Project:         "testproject",
		FileType:        "testfeed",
		FileDescription: "Test comment",
		Comments:        []string{"time of exp", "a good value"},
		Units:           []string{"NA", "parsecs"},
		Types:           []string{"time", "float"},
		Headers:         []string{"time", "testcol"},
	}
	assert.Equal(wantHeader, d.Headers()["testfeed"], "header with all sections")
}
