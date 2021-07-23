package parse

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/ctberthiaume/tsdata"
)

// Data holds parsed data from a single line of an underway feed
type Data struct {
	Feed      string
	Time      time.Time
	Throttled bool
	Values    []string
}

func (d Data) String() string {
	s := d.Feed
	if d.Throttled {
		s += "(T)"
	}
	s += ": "
	if !d.Time.IsZero() {
		s += d.Time.Format(time.RFC3339Nano)
	}
	if d.Values != nil {
		s = s + "," + strings.Join(d.Values, ",")
	}
	return s
}

// Line creates a delimited line of text, starting with RFC3339 timestamp.
func (d Data) Line(sep string) string {
	s := append([]string{d.Time.Format(time.RFC3339Nano)}, d.Values...)
	return strings.Join(s, sep)
}

// OK indicates whether this Data is complete (has a feed name and timestamp)
// and is not throttled.
func (d Data) OK() bool {
	return d.Feed != "" && !d.Time.IsZero() && !d.Throttled
}

// FeedCollection describes a multi-feed time-series data set.
type FeedCollection struct {
	Feeds map[string]tsdata.Tsdata
}

// NewFeedCollection creates a new FeedCollection struct
func NewFeedCollection() FeedCollection {
	d := FeedCollection{}
	d.Feeds = make(map[string]tsdata.Tsdata)
	return d
}

// Headers returns TSData headers for all feeds.
func (d FeedCollection) Headers() map[string]string {
	h := make(map[string]string)
	for feed, data := range d.Feeds {
		h[feed] = data.Header()
	}
	return h
}

// GeoThermDef defines where to find the most basic oceanographic underway data
// in a cruisemic output folder
type GeoThermDef struct {
	GeoFeed        string
	LatitudeCol    string
	LongitudeCol   string
	ThermoFeed     string
	TemperatureCol string
	SalinityCol    string
}

func (g GeoThermDef) GeoThermDefString() string {
	s, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}
	return string(s)
}
