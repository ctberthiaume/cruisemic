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
	i   int // line number in current stanza, e.g. $SEAFLOW is 1, geo is 2
	now func() time.Time
	DataManager
}

// NewGradients4Parser returns a pointer to a Gradients4Parser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewGradients4Parser(project string, interval time.Duration, now func() time.Time) Parser {
	metadata := tsdata.Tsdata{
		Project:         project,
		FileType:        "geo",
		FileDescription: "Gradients 4 Thompson underway feed",
		Comments:        []string{"RFC3339", "Latitude Decimal format", "Longitude Decimal format", "TSG temperature", "TSG conductivity", "TSG salinity"},
		Types:           []string{"time", "float", "float", "float", "float", "float"},
		Units:           []string{"NA", "deg", "deg", "C", "S/m", "PSU"},
		Headers:         []string{"time", "lat", "lon", "temp", "conductivity", "salinity"},
	}
	return &Gradients4Parser{
		Throttle:    NewThrottle(interval),
		now:         now,
		DataManager: *NewDataManager(metadata, interval),
	}
}

// ParseLine parses a single underway feed line. Only lines ending with \n are
// examined.
func (p *Gradients4Parser) ParseLine(line string) (d Data) {
	if len(line) == 0 || line[len(line)-1] != '\n' {
		return
	}

	// Remove trailing \n for parsing
	line = line[:len(line)-1]

	if line == "$SEAFLOW" {
		d = p.GetData()
		// Reset state beyond DataManager reset
		p.i = 0
		p.t = p.now().UTC()
		return
	}

	p.i++

	// Trim leading and trailing whitespace
	clean := strings.TrimSpace(line)

	switch {
	case p.i == 1:
		// Latitude
		if len(clean) < 2 {
			p.AddError(fmt.Errorf("Gradients4Parser: bad GPGGA latitude: line=%q", line))
		} else {
			latdd, latddErr := geo.GGALat2DD(clean[:len(clean)-1], clean[len(clean)-1:])
			if latddErr != nil {
				p.AddError(fmt.Errorf("Gradients4Parser: bad GPGGA lat: %v: line=%q", latddErr, line))
			} else {
				p.AddValue("lat", latdd)
			}
		}
	case p.i == 2:
		// Longitude
		if len(clean) < 2 {
			p.AddError(fmt.Errorf("Gradients4Parser: bad GPGGA longitude: line=%q", line))
		} else {
			londd, londdErr := geo.GGALon2DD(clean[:len(clean)-1], clean[len(clean)-1:])
			if londdErr != nil {
				p.AddError(fmt.Errorf("Gradients4Parser: bad GPGGA lon: %v: line=%q", londdErr, line))
			} else {
				p.AddValue("lon", londd)
			}
		}
	case p.i == 3:
		// Temperature
		_, floatErr := strconv.ParseFloat(clean, 64)
		if floatErr != nil {
			p.AddError(fmt.Errorf("Gradients4Parser: bad float: line=%q", line))
			p.AddValue("temp", tsdata.NA)
		} else {
			p.AddValue("temp", clean)
		}
	case p.i == 4:
		// Conductivity
		_, floatErr := strconv.ParseFloat(clean, 64)
		if floatErr != nil {
			p.AddError(fmt.Errorf("Gradients4Parser: bad float: line=%q", line))
			p.AddValue("conductivity", tsdata.NA)
		} else {
			p.AddValue("conductivity", clean)
		}
	case p.i == 5:
		// Salinity
		_, floatErr := strconv.ParseFloat(clean, 64)
		if floatErr != nil {
			p.AddError(fmt.Errorf("Gradients4Parser: bad float: line=%q", line))
			p.AddValue("salinity", tsdata.NA)
		} else {
			p.AddValue("salinity", clean)
		}
	}

	return
}
