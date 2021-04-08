package parse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testKMLineData struct {
	name        string
	input       string
	expected    Data
	expectError bool
}

func TestKMParserRegistry(t *testing.T) {
	assert := assert.New(t)
	constructor, ok := ParserRegistry["Kilo Moana"]
	assert.True(ok, "Kilo Moana parser is registered")
	if ok {
		p := constructor("testproject", 0)
		_, ok = p.(*KiloMoanaParser)
		assert.True(ok, "Kilo Moana parser is registered")
	}
}

func TestKMLines(t *testing.T) {
	t0 := time.Date(2017, 6, 17, 0, 30, 29, 365000000, time.UTC)
	testData := []testKMLineData{
		{"empty line", "", Data{}, false},
		{"bad feed ID", "2017 168 00 30 29 365 flr 78.000000", Data{}, false},
		{"fluor: good feed", "2017 168 00 30 29 365 flor 78.000000", Data{Feed: "fluor", Time: t0, Values: []string{"78.000000"}}, false},
		{"fluor: bad float", "2017 168 00 30 29 365 flor 78a.000000", Data{}, true},
		{"fluor: bad column count", "2017 168 00 30 29 365 flor 78.000000 foo", Data{}, true},
		{"fluor: bad date", "2a017 168 00 30 29 365 flor 78.000000", Data{}, true},
		{"thermo: good feed", "2017 168 00 30 29 365 uthsl 19.968599 0.040550 0.217500 27.397800", Data{Feed: "thermo", Time: t0, Values: []string{"19.968599", "0.040550", "0.217500", "27.397800"}}, false},
		{"thermo: bad float", "2017 168 00 30 29 365 uthsl 19.96a8599 0.040550 0.217500 27.397800", Data{}, true},
		{"thermo: bad column count", "2017 168 00 30 29 365 uthsl 19.968599 0.040550 0.217500 27.397800 foo", Data{}, true},
		{"thermo: bad date", "2017 16a8 00 30 29 365 uthsl 19.968599 0.040550 0.217500 27.397800", Data{}, true},
		{"geo: good feed", "$GPGGA,003029.00,2118.9043,N,15752.6526,W,2,7,0.8,27,M,,M,,*78", Data{Feed: "geo", Values: []string{"21.3151", "-157.8775"}}, false},
		{"geo: bad lat", "$GPGGA,003029.00,2198.9043,N,15752.6526,W,2,7,0.8,27,M,,M,,*78", Data{}, true},
		{"geo: bad lon", "$GPGGA,003029.00,2118.9043,N,15782.6526,W,2,7,0.8,27,M,,M,,*78", Data{}, true},
		{"geo: bad column count", "$GPGGA,003029.00,2118.9043,N,15752.6526,W,2,7,0.8,27,M,,M,,*78,foo", Data{}, true},
		{"heading: good feed", "$GPVTG,47.3,T,37.7,M,0.0,N,0.0,K,D*25", Data{Feed: "heading", Values: []string{"47.3", "0.0"}}, false},
		{"heading: bad track", "$GPVTG,47a.3,T,37.7,M,0.0,N,0.0,K,D*25", Data{}, true},
		{"heading: bad knots", "$GPVTG,47.3,T,37.7,M,0a.0,N,0.0,K,D*25", Data{}, true},
		{"heading: bad column count", "$GPVTG,47.3,T,37.7,M,0.0,N,0.0,K", Data{}, true},
		{"par: good feed", "2017 168 00 30 29 365 met  0.000 28.680  50.900 28.470 24.766  3.758 -0.246  1.097  1.099  0.000 5040.000  1.016 11.9 235.0 11.9   83.3 R-  0.000  0.000", Data{Feed: "par", Time: t0, Values: []string{"1.016"}}, false},
		{"par: good feed, no R-", "2017 168 00 30 29 365 met  0.000 28.680  50.900 28.470 24.766  3.758 -0.246  1.097  1.099  0.000 5040.000  1.016 11.9 235.0 11.9   83.3    0.000  0.000", Data{Feed: "par", Time: t0, Values: []string{"1.016"}}, false},
		{"par: bad float", "2017 168 00 30 29 365 met  0.000 28.680  50.900 28.470 24.766  3.758 -0.246  1.097  1.099  0.000 5040.000  1.a016 11.9 235.0 11.9   83.3 R-  0.000  0.000", Data{}, true},
		{"par: too few columns", "2017 168 00 30 29 365 met  0.000 28.680  50.900 28.470 24.766  3.758 -0.246  1.097  1.099  0.000 5040.000", Data{}, true},
		{"par: bad date", "2017 168 00 30 29 3a65 met  0.000 28.680  50.900 28.470 24.766  3.758 -0.246  1.097  1.099  0.000 5040.000  1.016 11.9 235.0 11.9   83.3 R-  0.000  0.000", Data{}, true},
	}
	for _, tt := range testData {
		t.Run(tt.name, createKMLinesTest(t, tt))
	}
}

func createKMLinesTest(t *testing.T, tt testKMLineData) func(*testing.T) {
	assert := assert.New(t)
	return func(t *testing.T) {
		p := NewKiloMoanaParser("test", 0)
		actual, err := p.ParseLine(tt.input)
		if !tt.expectError {
			assert.Nil(err, tt.name)
		} else {
			assert.NotNil(err, tt.name)
		}
		assert.Equal(tt.expected.Feed, actual.Feed, tt.name)
		assert.Equal(tt.expected.Values, actual.Values, tt.name)
		assert.Equal(tt.expected.Time.Format(time.RFC3339Nano), actual.Time.Format(time.RFC3339Nano), tt.name)
	}
}

func TestKMTimeOnUntimedLines(t *testing.T) {
	assert := assert.New(t)
	p := NewKiloMoanaParser("test", 0)
	// Parse a good thermo line to set last seen time. This should be the
	// time returned when parsing Geo and Heading
	_, _ = p.ParseLine("2017 168 00 30 29 365 uthsl 19.968599 0.040550 0.217500 27.397800")
	t0 := time.Date(2017, 6, 17, 0, 30, 29, 365000000, time.UTC)

	geo, err := p.ParseLine("$GPGGA,003029.00,2118.9043,N,15752.6526,W,2,7,0.8,27,M,,M,,*78")
	assert.Nil(err, "geo time")
	assert.Equal(t0, geo.Time, "geo time")

	heading, err := p.ParseLine("$GPVTG,47.3,T,37.7,M,0.0,N,0.0,K,D*25")
	assert.Nil(err, "heading time")
	assert.Equal(t0, heading.Time, "heading time")
}
