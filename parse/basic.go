package parse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctberthiaume/tsdata"
)

// BasicParser is the simplets feed parser.
type BasicParser struct {
	FeedCollection
	Throttle
	GeoThermDef
}

// NewBasicParser creates a pointer to a BasicParser.
func NewBasicParser(project string, interval time.Duration) Parser {
	p := &BasicParser{
		FeedCollection: NewFeedCollection(),
		Throttle:       NewThrottle(interval),
	}
	p.FeedCollection.Feeds["feed1"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "Basic-feed1",
		FileDescription: "Feed One",
		Types:           []string{"time", "float"},
		Units:           []string{"NA", "NA"},
		Headers:         []string{"time", "value1"},
	}
	p.FeedCollection.Feeds["feed2"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "Basic-feed2",
		FileDescription: "Feed Two",
		Types:           []string{"time", "float"},
		Units:           []string{"NA", "NA"},
		Headers:         []string{"time", "value2"},
	}
	return p
}

// ParseLine converts each feed line into a Data. Expects two whitespace
// delimited columns. The first should be RFC3339 formatted datetime. The
// second should be a floating point number. All Data returned have a feed type
// of "basic".
func (p *BasicParser) ParseLine(line string) (d Data, err error) {
	if len(line) == 0 {
		return
	}
	fields := strings.Fields(line)
	switch {
	case len(fields) == 3 && fields[2] == "feed1":
		if d, err = p.parseFeed1(fields); err != nil {
			return d, fmt.Errorf("BasicParser: bad feed1 line: %v: line=%q", err, line)
		}
	case len(fields) == 2 && fields[1] == "feed2":
		if d, err = p.parseFeed2(fields); err != nil {
			return d, fmt.Errorf("BasicParser: bad feed2 line: %v: line=%q", err, line)
		}
	}
	p.Limit(&d)
	return d, nil
}

func (p *BasicParser) parseFeed1(fields []string) (d Data, err error) {
	if len(fields) != 3 {
		return d, fmt.Errorf("incorrect field count %d", len(fields))
	}
	if _, err := strconv.ParseFloat(fields[1], 64); err != nil {
		return d, err
	}
	t, err := time.Parse(time.RFC3339, fields[0])
	if err != nil {
		return d, fmt.Errorf("bad date fields")
	}
	d.Feed = "feed1"
	d.Time = t
	d.Values = []string{fields[1]}
	return d, nil
}

func (p *BasicParser) parseFeed2(fields []string) (d Data, err error) {
	if len(fields) != 2 {
		return d, fmt.Errorf("incorrect field count %d", len(fields))
	}
	if _, err := strconv.ParseFloat(fields[0], 64); err != nil {
		return d, err
	}
	// This feed has no timestamp, so use the most recent from all feeds. If
	// there is no recent time, this can be caught by checking for
	// d.Time.IsZero() or d.OK().
	d.Feed = "feed2"
	d.Time = p.Recent("")
	d.Values = []string{fields[0]}
	return d, nil
}
