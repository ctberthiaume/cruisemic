package parse

import (
	"strings"
	"testing"
	"time"

	"github.com/ctberthiaume/cruisemic/storage"
	"github.com/stretchr/testify/assert"
)

type testG5LineData struct {
	name     string
	input    string
	expected map[string][]string
}

func TestG5ParserRegistry(t *testing.T) {
	assert := assert.New(t)
	constructor, ok := ParserRegistry["Gradients5"]
	assert.True(ok, "Gradients5 parser is registered")
	if ok {
		p := constructor("testproject", 0, time.Now)
		_, ok = p.(*Gradients5Parser)
		assert.True(ok, "Gradients5 parser is registered")
	}
}

func TestG5Lines(t *testing.T) {
	testData := []testG5LineData{
		{
			"good line",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64868,  31.2816::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\t3.64868\t31.2816\t157.580\n"},
			},
		},
		{
			"2 good stanzas",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64868,  31.2816::$PPAR, 157.580, 6.10, 5
$SEAFLOW::$GPZDA,213310.00,12,01,2023,00,00*6D::$GPGGA,213310.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3720,  3.64869,  31.2817::$PPAR, 158.580, 6.10, 5`,
			map[string][]string{
				"geo": {
					"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\t3.64868\t31.2816\t157.580\n",
					"2023-01-12T21:33:10Z\t47.6497\t-122.3134\t12.3720\t3.64869\t31.2817\t158.580\n",
				},
			},
		},
		{
			"2 good stanzas, with empty lines in between",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5

			$SEAFLOW::$GPZDA,213310.00,12,01,2023,00,00*6D::$GPGGA,213310.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 158.580, 6.10, 5`,
			map[string][]string{
				"geo": {
					"2023-01-12T21:33:09Z\t47.6497\t-122.3134\tNA\tNA\tNA\t157.580\n",
					"2023-01-12T21:33:10Z\t47.6497\t-122.3134\tNA\tNA\tNA\t158.580\n",
				},
			},
		},
		{
			"wrong number of fields",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{},
		},
		{
			"bad timestamp",
			`$SEAFLOW::$GPZDA,213a309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{},
		},
		{
			"bad lon",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,1f2218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{},
		},
		{
			"bad lat",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,47a38.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{},
		},
		{
			"bad lon direction",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,D,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{},
		},
		{
			"bad lat direction",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,P,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{},
		},
		{
			"empty TSG",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\tNA\tNA\tNA\t157.580\n"},
			},
		},
		{
			"bad temp",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.371a9,  3.64868,  31.2816::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\tNA\t3.64868\t31.2816\t157.580\n"},
			},
		},
		{
			"bad conductivity",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64a868,  31.2816::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\tNA\t31.2816\t157.580\n"},
			},
		},
		{
			"bad salinity",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64868,  31.2a816::$PPAR, 157.580, 6.10, 5`,
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\t3.64868\tNA\t157.580\n"},
			},
		},
		{
			"bad PAR, truncated",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 157.58, 6.10, 5`,
			map[string][]string{},
		},
		{
			"bad PAR, not a number",
			`$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44::::$PPAR, 1a7.580, 6.10, 5`,
			map[string][]string{},
		},
		{
			"missing PAR text entirely",
			"$SEAFLOW::$GPZDA,213309.00,12,01,2023,00,00*6D::$GPGGA,213309.00,4738.983141,N,12218.805824,W,2,17,0.7,15.773,M,-22.2,M,7.0,0402*44:: 12.3719,  3.64868,  31.2816::",
			map[string][]string{
				"geo": {"2023-01-12T21:33:09Z\t47.6497\t-122.3134\t12.3719\t3.64868\t31.2816\tNA\n"},
			},
		},
	}
	for _, tt := range testData {
		t.Run(tt.name, createG5LinesTest(t, tt))
	}
}

func createG5LinesTest(t *testing.T, tt testG5LineData) func(*testing.T) {
	assert := assert.New(t)

	return func(t *testing.T) {
		p := NewGradients5Parser("test", 0, time.Now)
		store, _ := storage.NewMemStorage()
		r := strings.NewReader(tt.input)
		err := ParseLines(p, r, store, false, true, false)
		assert.Nil(err, "writing for test: "+tt.name)
		// No need to check the raw feed
		// delete(store.Feeds, "raw")

		assert.Equal(tt.expected, store.Feeds, tt.name)
	}
}
