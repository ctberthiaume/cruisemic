package parse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/geo"
	"github.com/ctberthiaume/tsdata"
)

// KiloMoanaParser is a parser for Kilo Moana underway feed lines.
type KiloMoanaParser struct {
	Throttle
	t        time.Time         // time of current stanza
	values   map[string]string // accumulated values for this stanza by column name
	errors   []error           // errors for the current stanza
	metadata tsdata.Tsdata
}

// NewKiloMoanaParser returns a pointer to a KiloMoanaParser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewKiloMoanaParser(project string, interval time.Duration, now func() time.Time) Parser {
	_ = now // now is not sued in this function

	p := &KiloMoanaParser{
		Throttle: NewThrottle(interval),
		values:   make(map[string]string),
	}

	p.metadata = tsdata.Tsdata{
		Project:         project,
		FileType:        "geo",
		FileDescription: "Kilo Moana underway feed",
		Comments: []string{
			"RFC3339",
			"Thermosalinograph Temperature at Main Lab",
			"Thermosalinograph Conductivity",
			"Thermosalinograph Salinity",
			"Thermosalinograph Temperature at Bow",
			"Ship's Course (GPS COG)",
			"Ship's Speed (GPS SOG)",
			"Fluorometer raw scale count",
			"Surface PAR milliVolts",
			"Latitude Decimal format",
			"Longitude Decimal format",
		},
		Types:   []string{"time", "float", "float", "float", "float", "float", "float", "float", "float", "float", "float"},
		Units:   []string{"NA", "C", "S/m", "PSU", "C", "deg", "kn", "count", "mV", "deg", "deg"},
		Headers: []string{"time", "lab_temp", "conductivity", "salinity", "temp", "heading_true_north", "knots", "fluor", "par", "lat", "lon"},
	}

	return p
}

// ParseLine parses a single underway feed line. Only lines ending with \n are
// examined.
func (p *KiloMoanaParser) ParseLine(line string) (d Data) {
	if len(line) == 0 || line[len(line)-1] != '\n' {
		return
	}

	// Remove trailing \n for parsing
	line = line[:len(line)-1]

	var thisErr error
	if strings.HasPrefix(line, "$GPGGA") {
		fields := strings.Split(line, ",")
		if thisErr = p.parseGeo(fields); thisErr != nil {
			p.errors = append(p.errors, fmt.Errorf("KiloMoanaParser: bad GPGGA: %v: line=%q", thisErr, line))
		}
	} else if strings.HasPrefix(line, "$GPVTG") {
		fields := strings.Split(line, ",")
		if thisErr = p.parseHeading(fields); thisErr != nil {
			p.errors = append(p.errors, fmt.Errorf("KiloMoanaParser: bad GPVTG: %v: line=%q", thisErr, line))
		}
	} else {
		fields := strings.Fields(line)
		switch {
		case len(fields) >= 7 && fields[6] == "flor":
			if thisErr = p.parseFluor(fields); thisErr != nil {
				p.errors = append(p.errors, fmt.Errorf("KiloMoanaParser: bad fluor: %v: line=%q", thisErr, line))
			}
		case len(fields) >= 7 && fields[6] == "met":
			if thisErr = p.parsePar(fields); thisErr != nil {
				p.errors = append(p.errors, fmt.Errorf("KiloMoanaParser: bad met: %v: line=%q", thisErr, line))
			}
		case len(fields) >= 7 && fields[6] == "uthsl":
			if thisErr = p.parseThermo(fields); thisErr != nil {
				p.errors = append(p.errors, fmt.Errorf("KiloMoanaParser: bad uthsl: %v: line=%q", thisErr, line))
			}
		case len(fields) >= 7 && fields[6] == "bar1":
			d = p.createLastStanzaData()
			p.reset()

			// Start parsing the next stanza
			if thisErr = p.parseDate(fields); thisErr != nil {
				p.errors = append(p.errors, fmt.Errorf("KiloMoanaParser: bad bar1 date: %v: line=%q", thisErr, line))
			}
		}
	}

	return d
}

// Header returns a string header for a TSDATA file.
func (p *KiloMoanaParser) Header() string {
	return p.metadata.Header()
}

func (p *KiloMoanaParser) parseDate(fields []string) (err error) {
	if len(fields) < 6 {
		return fmt.Errorf("bad date fields")
	}
	parts := make([]int, 6)
	for i, f := range fields[:6] {
		parts[i], err = strconv.Atoi(f)
		if err != nil {
			return fmt.Errorf("bad date fields")
		}
	}
	t0 := time.Date(parts[0], time.January, 1, parts[2], parts[3], parts[4], parts[5]*1000000, time.UTC)
	p.t = t0.Add(time.Duration(24*(parts[1]-1)) * time.Duration(time.Hour)).Round(0)
	return
}

func (p *KiloMoanaParser) parseFluor(fields []string) (err error) {
	if len(fields) != 8 {
		return fmt.Errorf("incorrect field count %d", len(fields))
	}
	if _, err := strconv.ParseFloat(fields[7], 64); err != nil {
		return err
	}
	p.values["fluor"] = fields[7]
	return
}

func (p *KiloMoanaParser) parsePar(fields []string) (err error) {
	// This feed is space separated with padding for alignment, and
	// sometimes column 24 comes and goes, either blank or R-. These
	// two factors make column counting difficult. Rather than count
	// columns, just make sure there are at least 19 columns since we
	// only need columns 1-6 and 19.
	if len(fields) < 19 {
		return fmt.Errorf("incorrect field count %d", len(fields))
	}
	if _, err := strconv.ParseFloat(fields[18], 64); err != nil {
		return err
	}
	p.values["par"] = fields[18]
	return
}

func (p *KiloMoanaParser) parseThermo(fields []string) (err error) {
	if len(fields) != 11 {
		return fmt.Errorf("incorrect field count %d", len(fields))
	}
	for _, v := range fields[7:11] {
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return err
		}
	}
	p.values["lab_temp"] = fields[7]
	p.values["conductivity"] = fields[8]
	p.values["salinity"] = fields[9]
	p.values["temp"] = fields[10]
	return
}

func (p *KiloMoanaParser) parseGeo(fields []string) (err error) {
	if len(fields) != 15 {
		return fmt.Errorf("incorrect field count %d", len(fields))
	}
	latdd, latdderr := geo.GGALat2DD(fields[2], fields[3])
	if latdderr != nil {
		return latdderr
	}
	londd, londderr := geo.GGALon2DD(fields[4], fields[5])
	if londderr != nil {
		return londderr
	}
	p.values["lat"] = latdd
	p.values["lon"] = londd
	return
}

func (p *KiloMoanaParser) parseHeading(fields []string) (err error) {
	if len(fields) != 10 {
		return fmt.Errorf("incorrect field count %d", len(fields))
	}
	if _, err := strconv.ParseFloat(fields[1], 64); err != nil { // track
		return err
	}
	if _, err := strconv.ParseFloat(fields[5], 64); err != nil { // knots
		return err
	}
	p.values["heading_true_north"] = fields[1]
	p.values["knots"] = fields[5]
	return
}

// createLastStanzaData creates a completed Data struct once a stanza is finished.
func (p *KiloMoanaParser) createLastStanzaData() (d Data) {
	// Add errors regardless of whether the stanza has lat/lon/time
	d.Errors = p.errors

	_, latOk := p.values["lat"]
	_, lonOk := p.values["lon"]
	if latOk && lonOk {
		// Prepare complete Data struct
		d.Time = p.t
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
	}

	return d
}

func (p *KiloMoanaParser) reset() {
	p.t = time.Time{}
	p.values = make(map[string]string)
	p.errors = []error{}
}
