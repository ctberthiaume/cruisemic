package parse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/geo"
	"github.com/ctberthiaume/tsdata"
)

// SallyRideParser is a parser for Sally Ride underway feed lines.
type SallyRideParser struct {
	FeedCollection
	Throttle
	GeoThermDef
}

// NewSallyRideParser returns a pointer to a SallyRideParser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewSallyRideParser(project string, interval time.Duration) Parser {
	p := &SallyRideParser{
		FeedCollection: NewFeedCollection(),
		Throttle:       NewThrottle(interval),
	}
	p.FeedCollection.Feeds["wicor"] = tsdata.Tsdata{
		Project:         project,
		FileType:        "wicor-geo",
		FileDescription: "Sally Ride WICOR feed",
		Comments:        []string{"RFC3339", "Surface PAR", "Latitude Decimal format", "Longitude Decimal format", "Ship's Course (GPS COG)", "Ship's Speed (GPS SOG)", "Thermosalinograph Temperature at Bow", "Thermosalinograph Conductivity (milliSiemens per centimeter with slope, offset correction)", "Salinity", "Thermosalinograph Temperature at Main Lab", "Fluorometer"},
		Types:           []string{"time", "float", "float", "float", "float", "float", "float", "float", "float", "float", "float"},
		Units:           []string{"NA", "uE/s/m^2", "deg", "deg", "deg", "kn", "C", "mS/cm", "PSU", "C", "ug/L"},
		Headers:         []string{"time", "par", "lat", "lon", "heading", "speed", "bow_temp", "conductivity", "salinity", "lab_temp", "fluor"},
	}

	p.GeoThermDef = GeoThermDef{
		GeoFeed:        "wicor",
		LatitudeCol:    "lat",
		LongitudeCol:   "lon",
		ThermoFeed:     "wicor",
		TemperatureCol: "bow_temp",
		SalinityCol:    "salinity",
	}

	return p
}

// ParseLine parses and saves a single underway feed line.
func (p *SallyRideParser) ParseLine(line string) (d Data, err error) {
	var t time.Time
	if len(line) == 0 {
		return
	}
	if !strings.HasPrefix(line, "$WICOR") {
		return
	}
	fields := strings.Split(line, ",")
	if len(fields) < 7 {
		return
	}
	v := make(map[string]string)
FIELDSCAN:
	for i := 6; i < len(fields); i += 2 {
		code := fields[i]
		switch code {
		case "PA2", "CR1", "SP1", "TC2", "SA2", "FL2":
			// PAR, course, speed, conductivity, salinity, fluorometer
			if _, ok := v[code]; !ok {
				// For all of these values only consider the first one we see
				_, err = strconv.ParseFloat(fields[i-1], 64)
				if err != nil {
					return d, fmt.Errorf("SallyRideParser: bad %v: field=%q: %v", code, fields[i-1], err)
				}
				v[code] = fields[i-1]
				if code == "FL2" {
					// FL2 should be last value we need, stop looking
					break FIELDSCAN
				}
			}
		case "LA1":
			// Latitude
			if _, ok := v[code]; !ok {
				// Only consider first latitude we see
				if err = geo.CheckLat(fields[i-1]); err != nil {
					return d, fmt.Errorf("SallyRideParser: bad lat: field=%v: %v", fields[i-1], err)
				}
				v[code] = fields[i-1]
			}
		case "LO1":
			// Longitude
			if _, ok := v[code]; !ok {
				// Only consider first longitude we see
				if err = geo.CheckLon(fields[i-1]); err != nil {
					return d, fmt.Errorf("SallyRideParser: bad lon: field=%v: %v", fields[i-1], err)
				}
				v[code] = fields[i-1]
			}
		case "ZD1":
			// Epoch seconds
			if _, ok := v[code]; !ok {
				// Only consider first timestamp we see
				stamp, err := strconv.ParseInt(fields[i-1], 0, 64)
				if err != nil {
					return d, fmt.Errorf("SallyRideParser: bad timestamp: field=%v: %v", fields[i-1], err)
				}
				t = time.Unix(stamp, 0).UTC()
				v[code] = fields[i-1]
			}
		case "TT2":
			_, err := strconv.ParseFloat(fields[i-1], 64)
			if err != nil {
				return d, fmt.Errorf("SallyRideParser: bad %v: field=%v: %v", code, fields[i-1], err)
			}
			_, ok := v["TT2-Bow"]
			if !ok {
				v["TT2-Bow"] = fields[i-1] // First TT2 is Bow temp
			} else {
				_, ok := v["TT2-Lab"]
				if ok {
					// This shouldn't happen, should break before seeing 3 TT2 values
					return d, fmt.Errorf("SallyRideParser: saw too many TT2 values")
				}
				v["TT2-Lab"] = fields[i-1] // Second TT2 is Lab temp
			}
		}
	}
	if len(v) == 11 { // make sure all fields are accounted for
		d.Feed = "wicor"
		d.Values = []string{v["PA2"], v["LA1"], v["LO1"], v["CR1"], v["SP1"], v["TT2-Bow"], v["TC2"], v["SA2"], v["TT2-Lab"], v["FL2"]}
		d.Time = t
	} else {
		return d, fmt.Errorf("SallyRideParser: missing fields for line at %v", t.Format(time.RFC3339Nano))
	}
	p.Limit(&d)
	return d, nil
}
