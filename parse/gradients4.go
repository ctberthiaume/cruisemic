package parse

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ctberthiaume/cruisemic/geo"
	"github.com/ctberthiaume/tsdata"
)

// Gradients4Parser is a parser for Gradients 4 Thompson underway feed lines.
type Gradients4Parser struct {
	FeedCollection
	Throttle
	GeoThermDef
	i      int       // line number in current stanza, e.g. $SEAFLOW is 1, geo is 2
	t      time.Time // time of current stanza
	values []string  // accumulated values for this stanza
}

// NewGradients4Parser returns a pointer to a Gradients4Parser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewGradients4Parser(project string, interval time.Duration) Parser {
	p := &Gradients4Parser{
		FeedCollection: NewFeedCollection(),
		Throttle:       NewThrottle(interval),
		values:         []string{},
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
func (p *Gradients4Parser) ParseLine(line string) (err error) {
	if line == "$SEAFLOW" {
		// Reset state in case this interrupted a previous incomplete stanza
		p.i = 1
		p.t = time.Now().UTC()
		p.values = []string{}
		return
	}

	if p.i == 0 {
		// Haven't started a new stanza with $SEAFLOW line yet
		return
	}

	p.i++

	switch {
	case p.i == 2:
		// Latitude
		if len(line) < 2 {
			err = fmt.Errorf("Gradients4Parser: bad GPGGA latitude: line=%q", line)
			p.values = append(p.values, "NA")
		} else {
			latdd, latddErr := geo.GGALat2DD(line[:len(line)-1], line[len(line)-1:])
			if latddErr != nil {
				err = fmt.Errorf("Gradients4Parser: bad GPGGA lat: %v: line=%q", latddErr, line)
				latdd = "NA"
			} else {
				p.values = append(p.values, latdd)
			}
		}
	case p.i == 3:
		// Longitude
		if len(line) < 2 {
			err = fmt.Errorf("Gradients4Parser: bad GPGGA longitude: line=%q", line)
			p.values = append(p.values, "NA")
		} else {
			londd, londdErr := geo.GGALon2DD(line[:len(line)-1], line[len(line)-1:])
			if londdErr != nil {
				err = fmt.Errorf("Gradients4Parser: bad GPGGA lon: %v: line=%q", londdErr, line)
				londd = "NA"
			} else {
				p.values = append(p.values, londd)
			}
		}
	default:
		// All other values
		_, floatErr := strconv.ParseFloat(line, 64)
		if floatErr != nil {
			err = fmt.Errorf("Gradients4Parser: bad float: line=%q", line)
			line = "NA"
		}
		p.values = append(p.values, line)
	}

	// Return last parsing error
	return
}

func (p *Gradients4Parser) Flush() (d Data) {
	// fmt.Printf("%v %v %v\n", p.t, len(p.values), p.values)
	if len(p.values) == 6 {
		// Prepare complete Data struct
		d.Feed = "geo"
		d.Time = p.t
		d.Values = p.values
		p.Limit(&d)

		// Reset state
		p.i = 0
		p.t = time.Now().UTC()
		p.values = []string{}
		// fmt.Printf("%v\n", d)
	}
	return d
}
