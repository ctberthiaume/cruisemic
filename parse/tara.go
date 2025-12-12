package parse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/geo"
	"github.com/ctberthiaume/tsdata"
)

// TARAParser is a parser for only NMEA GPGGA lines.
type TARAParser struct {
	Throttle
	t        time.Time         // latest time read
	values   map[string]string // latest values by column name
	errors   []error           // errors encountered when parsing latest values
	metadata tsdata.Tsdata
}

// NewTARAParser returns a pointer to a TARAParser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewTARAParser(project string, interval time.Duration, now func() time.Time) Parser {
	_ = now // now is not used in this function

	p := &TARAParser{
		Throttle: NewThrottle(interval),
		values:   make(map[string]string),
	}

	p.metadata = tsdata.Tsdata{
		Project:         project,
		FileType:        "geo",
		FileDescription: "TARA feed",
		Comments: []string{
			"RFC3339",
			"Latitude Decimal format",
			"Longitude Decimal format",
		},
		Types:   []string{"time", "float", "float"},
		Units:   []string{"NA", "deg", "deg"},
		Headers: []string{"time", "lat", "lon"},
	}

	return p
}

// ParseLine parses a single underway feed line. Only lines ending with \n are
// examined.
func (p *TARAParser) ParseLine(line string) (d Data) {
	// Discard empty or incomplete lines
	if len(line) == 0 || line[len(line)-1] != '\n' {
		return
	}

	// Remove trailing \n for parsing
	line = line[:len(line)-1]

	var thisErr error
	if strings.HasPrefix(line, "$GPRMC") {
		fields := strings.Split(line, ",")
		if thisErr = p.parseGPRMC(fields); thisErr != nil {
			p.errors = append(p.errors, fmt.Errorf("TARAParser: bad GPRMC: %v: line=%q", thisErr, line))
		}
		d = p.createData()
		p.reset() // reset parser state
	}

	return d
}

// Header returns a string header for a TSDATA file.
func (p *TARAParser) Header() string {
	return p.metadata.Header()
}

func (p *TARAParser) parseGPRMC(fields []string) (err error) {
	if len(fields) < 13 {
		return fmt.Errorf("bad GPRMC fields")
	}

	// Parse lat/lon
	latdd, latdderr := geo.GGALat2DD(fields[3], fields[4])
	if latdderr != nil {
		return latdderr
	}
	londd, londderr := geo.GGALon2DD(fields[5], fields[6])
	if londderr != nil {
		return londderr
	}

	// Parse date/time
	if len(fields[9]) != 6 || len(fields[1]) != 6 {
		return fmt.Errorf("bad GPRMC date/time")
	}
	dateFields := []string{
		"20" + fields[9][4:6], // year
		fields[9][2:4],        // month
		fields[9][0:2],        // day
		fields[1][0:2],        // hour
		fields[1][2:4],        // minute
		fields[1][4:6],        // second
	}
	dateVals := make([]int, 7)
	for i, f := range dateFields {
		dateVals[i], err = strconv.Atoi(f)
		if err != nil {
			return fmt.Errorf("bad GPRMC date/time")
		}
	}
	dateVals[6] = 0 // nanoseconds
	t := time.Date(
		dateVals[0],
		time.Month(dateVals[1]),
		dateVals[2],
		dateVals[3],
		dateVals[4],
		dateVals[5],
		dateVals[6],
		time.UTC,
	)
	if t.Year() != dateVals[0] || int(t.Month()) != dateVals[1] || t.Day() != dateVals[2] {
		return fmt.Errorf("bad GPRMC date/time")
	}
	if t.Hour() != dateVals[3] || t.Minute() != dateVals[4] || t.Second() != dateVals[5] {
		return fmt.Errorf("bad GPRMC date/time")
	}

	p.values["lat"] = latdd
	p.values["lon"] = londd
	p.t = t

	return
}

// createData creates a completed Data struct.
func (p *TARAParser) createData() (d Data) {
	// Add errors regardless of whether the stanza has lat/lon/time
	d.Errors = p.errors
	// Add time
	d.Time = p.t

	// Add non-time values and throttle/limit
	d.Values = make([]string, len(p.metadata.Headers)-1)
	for i, k := range p.metadata.Headers {
		if k != "time" {
			val, ok := p.values[k]
			if !ok {
				d.Values[i-1] = tsdata.NA
			} else {
				d.Values[i-1] = val
			}
		}
	}
	p.Limit(&d)

	return d
}

func (p *TARAParser) reset() {
	p.t = time.Time{}
	p.values = make(map[string]string)
	p.errors = []error{}
}
