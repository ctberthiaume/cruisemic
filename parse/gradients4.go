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
	Throttle
	i        int               // line number in current stanza, e.g. $SEAFLOW is 1, geo is 2
	t        time.Time         // time of current stanza
	values   map[string]string // accumulated values for this stanza by column name
	errors   []error
	now      func() time.Time
	metadata tsdata.Tsdata
}

// NewGradients4Parser returns a pointer to a Gradients4Parser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewGradients4Parser(project string, interval time.Duration, now func() time.Time) Parser {
	p := &Gradients4Parser{
		Throttle: NewThrottle(interval),
		values:   make(map[string]string),
		now:      now,
	}
	p.metadata = tsdata.Tsdata{
		Project:         project,
		FileType:        "geo",
		FileDescription: "Gradients 4 Thompson underway feed",
		Comments:        []string{"RFC3339", "Latitude Decimal format", "Longitude Decimal format", "TSG temperature", "TSG conductivity", "TSG salinity"},
		Types:           []string{"time", "float", "float", "float", "float", "float"},
		Units:           []string{"NA", "deg", "deg", "C", "S/m", "PSU"},
		Headers:         []string{"time", "lat", "lon", "temp", "conductivity", "salinity"},
	}

	return p
}

// ParseLine parses and saves a single underway feed line.
func (p *Gradients4Parser) ParseLine(line string) (d Data) {
	if line == "$SEAFLOW" {
		d = p.createLastStanzaData()
		p.reset()
		return
	}

	p.i++

	// Trim leading and trailing whitespace
	clean := strings.TrimSpace(line)

	switch {
	case p.i == 1:
		// Latitude
		if len(clean) < 2 {
			p.errors = append(p.errors, fmt.Errorf("Gradients4Parser: bad GPGGA latitude: line=%q", line))
		} else {
			latdd, latddErr := geo.GGALat2DD(clean[:len(clean)-1], clean[len(clean)-1:])
			if latddErr != nil {
				p.errors = append(p.errors, fmt.Errorf("Gradients4Parser: bad GPGGA lat: %v: line=%q", latddErr, line))
			} else {
				p.values["lat"] = latdd
			}
		}
	case p.i == 2:
		// Longitude
		if len(clean) < 2 {
			p.errors = append(p.errors, fmt.Errorf("Gradients4Parser: bad GPGGA longitude: line=%q", line))
		} else {
			londd, londdErr := geo.GGALon2DD(clean[:len(clean)-1], clean[len(clean)-1:])
			if londdErr != nil {
				p.errors = append(p.errors, fmt.Errorf("Gradients4Parser: bad GPGGA lon: %v: line=%q", londdErr, line))
			} else {
				p.values["lon"] = londd
			}
		}
	case p.i == 3:
		// Temperature
		_, floatErr := strconv.ParseFloat(clean, 64)
		if floatErr != nil {
			p.errors = append(p.errors, fmt.Errorf("Gradients4Parser: bad float: line=%q", line))
			p.values["temp"] = tsdata.NA
		} else {
			p.values["temp"] = clean
		}
	case p.i == 4:
		// Conductivity
		_, floatErr := strconv.ParseFloat(clean, 64)
		if floatErr != nil {
			p.errors = append(p.errors, fmt.Errorf("Gradients4Parser: bad float: line=%q", line))
			p.values["conductivity"] = tsdata.NA
		} else {
			p.values["conductivity"] = clean
		}
	case p.i == 5:
		// Salinity
		_, floatErr := strconv.ParseFloat(clean, 64)
		if floatErr != nil {
			p.errors = append(p.errors, fmt.Errorf("Gradients4Parser: bad float: line=%q", line))
			p.values["salinity"] = tsdata.NA
		} else {
			p.values["salinity"] = clean
		}
	}

	return
}

// Header returns a string header for a TSDATA file.
func (p *Gradients4Parser) Header() string {
	return p.metadata.Header()
}

func (p *Gradients4Parser) createLastStanzaData() (d Data) {
	// Add errors regardless of whether the stanza is complete
	d.Errors = p.errors

	if len(p.values) == 5 {
		// Prepare complete Data struct
		d.Time = p.t
		d.Values = make([]string, len(p.values))
		for i, k := range p.metadata.Headers {
			if k != "time" {
				d.Values[i-1] = p.values[k]
			}
		}
		p.Limit(&d)
	}

	return d
}

func (p *Gradients4Parser) reset() {
	// Reset state
	p.i = 0
	p.t = p.now().UTC()
	p.values = make(map[string]string)
	p.errors = []error{}
}
