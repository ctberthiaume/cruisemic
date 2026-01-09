package parse

import (
	"strings"
	"testing"
	"time"

	"github.com/ctberthiaume/cruisemic/storage"
	"github.com/stretchr/testify/assert"
)

type testTN448LineData struct {
	name     string
	input    string
	expected map[string][]string
}

func TestTN448ParserRegistry(t *testing.T) {
	assert := assert.New(t)
	constructor, ok := ParserRegistry["TN448"]
	assert.True(ok, "TN448 parser is registered")
	if ok {
		p := constructor("testproject", 0, time.Now)
		_, ok = p.(*TN448Parser)
		assert.True(ok, "TN448 parser is registered")
	}
}

func TestTN448Lines(t *testing.T) {
	testData := []testTN448LineData{
		{
			"TN448 line with extra tsg field",
			`$SEAFLOW::$GNZDA,213218.00,31,10,2023,00,00*6D::$GNGGA,213218.00,4737.578758,N,12222.827136,W,2,15,0.8,12.181,M,-22.0,M,4.0,0402*4F:: 15.0526,  3.78840,  30.4126, 1501.506::
`,
			map[string][]string{
				"geo": {"2023-10-31T21:32:18Z\t47.6263\t-122.3805\t15.0526\t3.78840\t30.4126\tNA\n"},
			},
		},
		{
			"good line",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64868,  31.2816::157.580
`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\t3.64868\t31.2816\t157.580\n"},
			},
		},
		{
			"2 good stanzas",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64868,  31.2816::157.580
$SEAFLOW::$GNZDA,213310.00,12,01,2023,00,00*6D::$GNGGA,213310.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3720,  3.64869,  31.2817::158.580
`,
			map[string][]string{
				"geo": {
					"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\t3.64868\t31.2816\t157.580\n",
					"2023-01-12T21:33:10Z\t47.6497\t-122.3134\t12.3720\t3.64869\t31.2817\t158.580\n",
				},
			},
		},
		{
			"2 good stanzas, with empty lines in between",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580

$SEAFLOW::$GNZDA,213310.00,12,01,2023,00,00*6D::$GNGGA,213310.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::158.580
`,
			map[string][]string{
				"geo": {
					"2023-01-12T21:33:09Z\t47.6497\t-122.3134\tNA\tNA\tNA\t157.580\n",
					"2023-01-12T21:33:10Z\t47.6497\t-122.3134\tNA\tNA\tNA\t158.580\n",
				},
			},
		},
		{
			"wrong number of fields",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::::157.580
`,
			map[string][]string{},
		},
		{
			"bad timestamp, not a number",
			`$SEAFLOW::$GNZDA,21a309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{},
		},
		{
			"bad timestamp, timestamp too long",
			`$SEAFLOW::$GNZDA,213309.001,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{},
		},
		{
			"bad timestamp, incorrect number of time fields",
			`$SEAFLOW::$GNZDA,213309.00,12,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{},
		},
		{
			"bad GNGGA, too many fields",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,X,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{},
		},
		{
			"bad lon",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,1f2218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{},
		},
		{
			"bad lat",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,47a38.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{},
		},
		{
			"bad lon direction",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,D,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{},
		},
		{
			"bad lat direction",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,P,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{},
		},
		{
			"empty TSG",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.580
`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\tNA\tNA\tNA\t157.580\n"},
			},
		},
		{
			"bad temp",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.371a9,  3.64868,  31.2816::157.580
`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\tNA\t3.64868\t31.2816\t157.580\n"},
			},
		},
		{
			"bad conductivity",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64a868,  31.2816::157.580
`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\tNA\t31.2816\t157.580\n"},
			},
		},
		{
			"bad salinity",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64868,  31.2a816::157.580
`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\t3.64868\tNA\t157.580\n"},
			},
		},
		{
			"bad PAR, truncated",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::157.58
`,
			map[string][]string{},
		},
		{
			"bad PAR, not a number",
			`$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::1a7.580
`,
			map[string][]string{},
		},
		{
			"missing PAR text entirely",
			"$SEAFLOW::$GNZDA,213309.00,12,01,2023,00,00*6D::$GNGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64868,  31.2816::\n",
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\t3.64868\t31.2816\tNA\n"},
			},
		},
	}
	for _, tt := range testData {
		t.Run(tt.name, createTN448LinesTest(t, tt))
	}
}

func createTN448LinesTest(t *testing.T, tt testTN448LineData) func(*testing.T) {
	assert := assert.New(t)

	return func(t *testing.T) {
		p := NewTN448Parser("test", 0, time.Now)
		store, _ := storage.NewMemStorage()
		r := strings.NewReader(tt.input)
		err := ParseLines(p, r, store, true, false)
		assert.Nil(err, "writing for test: "+tt.name)
		// No need to check the raw feed
		// delete(store.Feeds, "raw")

		assert.Equal(tt.expected, store.Feeds, tt.name)
	}
}
