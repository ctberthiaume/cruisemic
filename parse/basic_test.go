package parse

import (
	"bufio"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasicParserRegistry(t *testing.T) {
	assert := assert.New(t)
	constructor, ok := ParserRegistry["Basic"]
	assert.True(ok, "Basic parser is registered")
	if ok {
		p := constructor("testproject", 0)
		_, ok = p.(*BasicParser)
		assert.True(ok, "Basic parser is registered")
	}
}

type testBasicLineData struct {
	name        string
	input       string
	expected    Data
	expectError bool
}

func TestBasicLines(t *testing.T) {
	t0, _ := time.Parse(time.RFC3339, "2017-06-17T00:30:29.365000000Z")
	testData := []testBasicLineData{
		{"empty line", "", Data{}, false},
		{"bad feed1 ID", "2017-06-17T00:30:29.365000000Z 10 feedNone", Data{}, false},
		{"bad column count", "2017-06-17T00:30:29.365000000Z", Data{}, false},
		{"bad feed1 date", "201a7-06-17T00:30:29.365000000Z 10 feed1", Data{}, true},
		{"bad feed1 value", "201a7-06-17T00:30:29.365000000Z 1a0 feed1", Data{}, true},
		{
			"good line feed1",
			"2017-06-17T00:30:29.365000000Z 10 feed1",
			Data{Feed: "feed1", Time: t0, Values: []string{"10"}},
			false,
		},
		{"bad feed2 ID", "10 feed2None", Data{}, false},
		{"bad feed2 value", "1a0 feed2", Data{}, true},
		{
			"good line feed2",
			"10 feed2",
			Data{Feed: "feed2", Values: []string{"10"}},
			false,
		},
	}
	for _, tt := range testData {
		t.Run(tt.name, createBasicLinesTest(t, tt))
	}
}

func createBasicLinesTest(t *testing.T, tt testBasicLineData) func(*testing.T) {
	assert := assert.New(t)
	return func(t *testing.T) {
		p := NewBasicParser("basic", 0)
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

func TestBasicHeaders(t *testing.T) {
	assert := assert.New(t)
	p := NewBasicParser("test", 0)
	h := p.Headers()
	keys := make([]string, len(h))
	i := 0
	for k := range h {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	assert.Equal(keys, []string{"feed1", "feed2"})
}

func TestBasicTimeOnUntimedLines(t *testing.T) {
	assert := assert.New(t)
	p := NewBasicParser("test", 0)

	// Parse a feed2 line, should not have a time, but is otherwise parsed OK.
	d, err := p.ParseLine("10 feed2")
	assert.Nil(err, "first feed2 returns nil error")
	assert.Equal("feed2", d.Feed, "first feed2 has Feed")
	assert.Equal([]string{"10"}, d.Values, "first feed2 has Values")
	assert.True(d.Time.IsZero(), "first feed2 has no time")
	// Parse a good feed1 line to set last seen time. This should be the time
	// returned when parsing feed2 immediately after.
	t0, _ := time.Parse(time.RFC3339, "2017-06-17T00:30:29.365000000Z")
	_, _ = p.ParseLine("2017-06-17T00:30:29.365000000Z 10 feed1")
	d, err = p.ParseLine("10 feed2")
	assert.Nil(err, "second feed2 returns nil error")
	assert.Equal("feed2", d.Feed, "second feed2 has Feed")
	assert.Equal([]string{"10"}, d.Values, "second feed2 has Values")
	assert.Equal(t0.Format(time.RFC3339Nano), d.Time.Format(time.RFC3339Nano), "second feed2 has correct time")
}

var d Data

func BenchmarkBasicLines(b *testing.B) {
	// (10**9 / ns/op) should show lines/second in simplest case
	// On my laptop this is (10**9 / 1727 ns/op) == 579,038 lines / sec
	gen := lineGenerator(1)
	p := NewBasicParser("test", 0)
	for n := 0; n < b.N; n++ {
		b := []byte(gen())
		n := len(b)
		n = Whitelist(b, n)

		scanner := bufio.NewScanner(strings.NewReader(string(b[:n])))
		for scanner.Scan() {
			line := scanner.Text()
			d, _ = p.ParseLine(line)
			p.Limit(&d)
		}
	}
	fmt.Printf("%v\n", d.String())
}

func lineGenerator(tick float64) func() string {
	t, _ := time.Parse(time.RFC3339, "2000-01-01T00:00:00Z")
	s := 0. // seconds progress
	rand.Seed(time.Now().UnixNano())
	return func() string {
		t = t.Add(time.Duration(float64(time.Second) + s))
		s += tick
		if rand.Intn(2) == 1 {
			tstr := t.Format(time.RFC3339Nano)
			return fmt.Sprintf("%v %v feed1\r\n", tstr, rand.Int31n(100))
		}
		return fmt.Sprintf("%v feed2\r\n", rand.Float32())
	}
}
