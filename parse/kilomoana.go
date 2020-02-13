package parse

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/geo"
	"github.com/ctberthiaume/tsdata"
)

// KiloMoanaParser is a parser for Kilo Moana underway feed lines.
type KiloMoanaParser struct {
	FeedCollection
	Throttle
}

// NewKiloMoanaParser returns a pointer to a KiloMoanaParser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewKiloMoanaParser(project string, interval time.Duration) Parser {
	p := &KiloMoanaParser{
		FeedCollection: NewFeedCollection(),
		Throttle:       NewThrottle(interval),
	}
	p.FeedCollection.Feeds["fluor"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "fluor",
		FileDescription: "Kilo Moana underway feed Fluorescence",
		Comments:        []string{"RFC3339", "Fluorometer raw scale count"},
		Types:           []string{"time", "float"},
		Units:           []string{"NA", "count"},
		Headers:         []string{"time", "fluor"},
	}
	p.FeedCollection.Feeds["geo"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "geo",
		FileDescription: "Kilo Moana underway feed decimal degree coordinates",
		Comments:        []string{"RFC3339", "Latitude Decimal format", "Longitude Decimal format"},
		Types:           []string{"time", "float", "float"},
		Units:           []string{"NA", "deg", "deg"},
		Headers:         []string{"time", "lat", "lon"},
	}
	p.FeedCollection.Feeds["heading"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "heading",
		FileDescription: "Kilo Moana underway feed cruise track direction and speed",
		Comments:        []string{"RFC3339", "Ship's Course (GPS COG)", "Ship's Speed (GPS SOG)"},
		Types:           []string{"time", "float", "float"},
		Units:           []string{"NA", "deg", "kn"},
		Headers:         []string{"time", "heading_true_north", "knots"},
	}
	p.FeedCollection.Feeds["par"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "par",
		FileDescription: "Kilo Moana underway feed PAR",
		Comments:        []string{"RFC3339", "Surface PAR milliVolts"},
		Types:           []string{"time", "float"},
		Units:           []string{"NA", "mV"},
		Headers:         []string{"time", "par"},
	}
	p.FeedCollection.Feeds["thermo"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "thermo",
		FileDescription: "Kilo Moana underway feed thermosalinograph",
		Comments:        []string{"RFC3339", "Thermosalinograph Temperature at Main Lab", "Thermosalinograph Conductivity", "Salinity", "Thermosalinograph Temperature at Bow"},
		Types:           []string{"time", "float", "float", "float", "float"},
		Units:           []string{"NA", "C", "S/m", "PSU", "C"},
		Headers:         []string{"time", "lab_temp", "conductivity", "salinity", "bow_temp"},
	}
	return p
}

// ParseLine parses and saves a single underway feed line.
func (p *KiloMoanaParser) ParseLine(line string) (d Data) {
	var err error
	if len(line) == 0 {
		return d
	}
	if strings.HasPrefix(line, "$GPGGA") {
		fields := strings.Split(line, ",")
		if d, err = p.parseGeo(fields); err != nil {
			log.Printf("format: bad GPGGA: %v: line=%q", err, line)
		}
	} else if strings.HasPrefix(line, "$GPVTG") {
		fields := strings.Split(line, ",")
		if d, err = p.parseHeading(fields); err != nil {
			log.Printf("format: bad GPVTG: %v: line=%q", err, line)
		}
	} else {
		fields := strings.Fields(line)
		switch {
		case len(fields) >= 7 && fields[6] == "flor":
			if d, err = p.parseFluor(fields); err != nil {
				log.Printf("format: bad fluor: %v: line=%q", err, line)
			}
		case len(fields) >= 7 && fields[6] == "met":
			if d, err = p.parsePar(fields); err != nil {
				log.Printf("format: bad met: %v: line=%q", err, line)
			}
		case len(fields) >= 7 && fields[6] == "uthsl":
			if d, err = p.parseThermo(fields); err != nil {
				log.Printf("format: bad uthsl: %v: line=%q", err, line)
			}
		}
	}
	p.Limit(&d)
	return d
}

func (p *KiloMoanaParser) parseDate(fields []string) (t time.Time, err error) {
	if len(fields) < 6 {
		return t, fmt.Errorf("bad date fields")
	}
	parts := make([]int, 6)
	for i, f := range fields[:6] {
		parts[i], err = strconv.Atoi(f)
		if err != nil {
			return t, fmt.Errorf("bad date fields")
		}
	}
	t0 := time.Date(parts[0], time.January, 1, parts[2], parts[3], parts[4], parts[5]*1000000, time.UTC)
	t = t0.Add(time.Duration(24*(parts[1]-1)) * time.Duration(time.Hour)).Round(0)
	return t, nil
}

func (p *KiloMoanaParser) parseFluor(fields []string) (d Data, err error) {
	if len(fields) != 8 {
		return d, fmt.Errorf("incorrect field count %d", len(fields))
	}
	if _, err := strconv.ParseFloat(fields[7], 64); err != nil {
		return d, err
	}
	t, err := p.parseDate(fields)
	if err != nil {
		return d, err
	}
	d.Feed = "fluor"
	d.Time = t
	d.Values = fields[7:8]
	return d, nil
}

func (p *KiloMoanaParser) parsePar(fields []string) (d Data, err error) {
	// This feed is space separated with padding for alignment, and
	// sometimes column 24 comes and goes, either blank or R-. These
	// two factors make column counting difficult. Rather than count
	// columns, just make sure there are at least 19 columns since we
	// only need columns 1-6 and 19.
	if len(fields) < 19 {
		return d, fmt.Errorf("incorrect field count %d", len(fields))
	}
	if _, err := strconv.ParseFloat(fields[18], 64); err != nil {
		return d, err
	}
	t, err := p.parseDate(fields)
	if err != nil {
		return d, err
	}
	d.Feed = "par"
	d.Time = t
	d.Values = fields[18:19]
	return d, nil
}

func (p *KiloMoanaParser) parseThermo(fields []string) (d Data, err error) {
	if len(fields) != 11 {
		return d, fmt.Errorf("incorrect field count %d", len(fields))
	}
	for _, v := range fields[7:11] {
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return d, err
		}
	}
	t, err := p.parseDate(fields)
	if err != nil {
		return d, err
	}
	d.Feed = "thermo"
	d.Time = t
	d.Values = fields[7:11]
	return d, nil
}

func (p *KiloMoanaParser) parseGeo(fields []string) (d Data, err error) {
	if len(fields) != 15 {
		return d, fmt.Errorf("incorrect field count %d", len(fields))
	}
	latdd, latdderr := geo.GGALat2DD(fields[2], fields[3])
	if latdderr != nil {
		return d, latdderr
	}
	londd, londderr := geo.GGALon2DD(fields[4], fields[5])
	if londderr != nil {
		return d, londderr
	}
	d.Feed = "geo"
	d.Time = p.Recent("thermo") // pin to thermosalinograph time
	d.Values = []string{latdd, londd}
	return d, nil
}

func (p *KiloMoanaParser) parseHeading(fields []string) (d Data, err error) {
	if len(fields) != 10 {
		return d, fmt.Errorf("incorrect field count %d", len(fields))
	}
	if _, err := strconv.ParseFloat(fields[1], 64); err != nil { // track
		return d, err
	}
	if _, err := strconv.ParseFloat(fields[5], 64); err != nil { // knots
		return d, err
	}
	d.Feed = "heading"
	d.Time = p.Recent("thermo") // pin to thermosalinograph time
	d.Values = []string{fields[1], fields[5]}
	return d, nil
}
