package parse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/geo"
	"github.com/ctberthiaume/tsdata"
)

// Gradients4Parser is a parser for Gradients 4 Thompson underway feed lines.
type Gradients4Parser struct {
	FeedCollection
	Throttle
	GeoThermDef
}

// NewGradients4Parser returns a pointer to a Gradients4Parser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewGradients4Parser(project string, interval time.Duration) Parser {
	p := &Gradients4Parser{
		FeedCollection: NewFeedCollection(),
		Throttle:       NewThrottle(interval),
	}
	p.FeedCollection.Feeds["geo"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "geo",
		FileDescription: "Gradients 4 Thompson underway feed",
		Comments:        []string{"RFC3339", "Latitude Decimal format", "Longitude Decimal format", "TSG temperature", "TSG conductivity", "TSG salinity", "PAR"},
		Types:           []string{"time", "float", "float", "float", "float", "float", "float"},
		Units:           []string{"NA", "deg", "deg", "C", "S/m", "PSU", "mV"},
		Headers:         []string{"time", "lat", "lon", "temp", "conductivity", "salinity", "par"},
	}

	p.GeoThermDef = GeoThermDef{
		GeoFeed:         "geo",
		LatitudeCol:     "lat",
		LongitudeCol:    "lon",
		ThermoFeed:      "geo",
		TemperatureCol:  "temp",
		SalinityCol:     "salinity",
		ConductivityCol: "conductivity",
	}

	return p
}

// ParseLine parses and saves a single underway feed line.
func (p *Gradients4Parser) ParseLine(line string) (d Data, err error) {
	if len(line) == 0 {
		return
	}

	t := time.Now().UTC()

	fields := strings.Fields(line)
	if len(fields) != 6 {
		return d, fmt.Errorf("Gradients4Parser: incorrect field count %d: line=%q: time=%s", len(fields), line, t.Format(time.RFC3339Nano))
	}

	latstr := strings.Replace(fields[0], ",", "", -1)
	lonstr := strings.Replace(fields[1], ",", "", -1)
	tempstr := strings.Replace(fields[2], ",", "", -1)
	condstr := strings.Replace(fields[3], ",", "", -1)
	salstr := strings.Replace(fields[4], ",", "", -1)
	parstr := strings.Replace(fields[5], ",", "", -1)

	// Latitude
	if err = geo.CheckLat(latstr); err != nil {
		return d, fmt.Errorf("Gradients4Parser: bad lat: field=%v: %v", latstr, err)
	}
	// Longitude
	if err = geo.CheckLon(lonstr); err != nil {
		return d, fmt.Errorf("Gradients4Parser: bad lon: field=%v: %v", lonstr, err)
	}
	// Temperature
	if _, err = strconv.ParseFloat(tempstr, 64); err != nil {
		return d, fmt.Errorf("Gradients4Parser: bad temp: field=%v: %v", tempstr, err)
	}
	// Conductivity
	if _, err = strconv.ParseFloat(condstr, 64); err != nil {
		return d, fmt.Errorf("Gradients4Parser: bad cond: field=%v: %v", condstr, err)
	}
	// Salinity
	if _, err = strconv.ParseFloat(salstr, 64); err != nil {
		return d, fmt.Errorf("Gradients4Parser: bad sal: field=%v: %v", salstr, err)
	}
	// PAR
	if _, err = strconv.ParseFloat(parstr, 64); err != nil {
		return d, fmt.Errorf("Gradients4Parser: bad par: field=%v: %v", parstr, err)
	}

	d.Feed = "geo"
	d.Values = []string{latstr, lonstr, tempstr, condstr, salstr, parstr}
	d.Time = t

	p.Limit(&d)
	return d, nil
}
