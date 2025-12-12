package parse

import (
	"strings"
	"testing"
	"time"

	"github.com/ctberthiaume/cruisemic/storage"
	"github.com/stretchr/testify/assert"
)

type testTARALineData struct {
	name     string
	input    string
	expected map[string][]string
}

func TestTARAParserRegistry(t *testing.T) {
	assert := assert.New(t)
	constructor, ok := ParserRegistry["TARA"]
	assert.True(ok, "TARA parser is registered")
	if ok {
		p := constructor("testproject", 0, time.Now)
		_, ok = p.(*TARAParser)
		assert.True(ok, "TARA parser is registered")
	}
}

func TestTARALines(t *testing.T) {
	testData := []testTARALineData{
		{
			"good TARA GPRMC line",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{
				"geo": {"2025-12-07T16:03:32Z\t47.7295\t-3.3740\n"},
			},
		},
		{
			"good TARA GPRMC line with carriage return",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\r\n",
			map[string][]string{
				"geo": {"2025-12-07T16:03:32Z\t47.7295\t-3.3740\n"},
			},
		},
		{
			"TARA GPRMC line with no newline",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19",
			map[string][]string{},
		},
		{
			"Partial TARA GPRMC line with malformed start",
			"RMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"Partial TARA GPRMC line with truncated ending",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,",
			map[string][]string{},
		},
		{
			"two good TARA GPRMC lines with empty lines between",
			`$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19


$GPRMC,170332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19
`,
			map[string][]string{
				"geo": {
					"2025-12-07T16:03:32Z\t47.7295\t-3.3740\n",
					"2025-12-07T17:03:32Z\t47.7295\t-3.3740\n"},
			},
		},
		{
			"two good TARA GPRMC lines with extra lines between",
			`$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19
$GPGGA,160333,4743.7696,N,00322.4403,W,2,09,1.6,-7.1,M,,M,,*66
$GPGLL,4743.7696,N,00322.4403,W,160333,A,D*5E
$GPRMC,170332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19
`,
			map[string][]string{
				"geo": {
					"2025-12-07T16:03:32Z\t47.7295\t-3.3740\n",
					"2025-12-07T17:03:32Z\t47.7295\t-3.3740\n"},
			},
		},
		{
			"bad date, day out of range",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,401425,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, day not a number",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,AB1425,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, month out of range",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,071425,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, month not a number",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,07AB25,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, year not a number",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,0712AB,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, too long",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,07122509,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, too short",
			"$GPRMC,160332,A,4743.7694,N,00322.4405,W,0.0,182.6,07122,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad time, hour out of range",
			"$GPRMC,300332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, hour not a number",
			"$GPRMC,AB0332,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad time, min out of range",
			"$GPRMC,166132,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, min not a number",
			"$GPRMC,16AB32,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad time, sec out of range",
			"$GPRMC,160361,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad date, sec not a number",
			"$GPRMC,1603AB,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad time, too short",
			"$GPRMC,16033,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad time, too long",
			"$GPRMC,1603322,A,4743.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad lat number",
			"$GPRMC,160332,A,47A3.7694,N,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad lon number",
			"$GPRMC,1603322,A,4743.7694,N,00A22.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad lat direction",
			"$GPRMC,160332,A,47A3.7694,X,00322.4405,W,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
		{
			"bad lon direction",
			"$GPRMC,1603322,A,4743.7694,N,00A22.4405,X,0.0,182.6,071225,0.2,W,D*19\n",
			map[string][]string{},
		},
	}
	for _, tt := range testData {
		t.Run(tt.name, createTARALinesTest(t, tt))
	}
}

func createTARALinesTest(t *testing.T, tt testTARALineData) func(*testing.T) {
	assert := assert.New(t)

	return func(t *testing.T) {
		p := NewTARAParser("test", 0, time.Now)
		store, _ := storage.NewMemStorage()
		r := strings.NewReader(tt.input)
		err := ParseLines(p, r, store, true, false)
		assert.Nil(err, "writing for test: "+tt.name)

		// No need to check the raw feed
		//delete(store.Feeds, "raw")

		assert.Equal(tt.expected, store.Feeds, tt.name)
	}
}
