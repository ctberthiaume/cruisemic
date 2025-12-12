package parse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/geo"
	"github.com/ctberthiaume/tsdata"
)

// TN427Parser is a parser for TN427 (and possibly TN428 ...) Thompson underway feed lines.
type TN427Parser struct {
	Throttle
	metadata tsdata.Tsdata
}

// NewTN427Parser returns a pointer to a TN427Parser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewTN427Parser(project string, interval time.Duration, now func() time.Time) Parser {
	p := &TN427Parser{
		Throttle: NewThrottle(interval),
	}
	p.metadata = tsdata.Tsdata{
		Project:         project,
		FileType:        "geo",
		FileDescription: "TN427+ Thompson underway feed",
		Comments:        []string{"RFC3339", "Latitude Decimal format", "Longitude Decimal format", "TSG temperature", "TSG conductivity", "TSG salinity", "PAR"},
		Types:           []string{"time", "float", "float", "float", "float", "float", "float"},
		Units:           []string{"NA", "deg", "deg", "C", "S/m", "PSU", "µE/m^2/s"},
		Headers:         []string{"time", "lat", "lon", "temp", "conductivity", "salinity", "par"},
	}

	return p
}

// ParseLine parses a single underway feed line. Only lines ending with \n are
// examined.
func (p *TN427Parser) ParseLine(line string) (d Data) {
	if len(line) == 0 || line[len(line)-1] != '\n' {
		return
	}

	// Remove trailing \n for parsing
	line = line[:len(line)-1]

	values := make(map[string]string)

	// Trim leading and trailing whitespace
	clean := strings.TrimSpace(line)

	if !strings.HasPrefix(clean, "$SEAFLOW") {
		return
	}

	fields := strings.Split(clean, "::")

	if len(fields) != 5 {
		return
	}

	// Parse time
	timeFields := strings.Split(fields[1], ",")
	if len(timeFields) != 7 {
		d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad GPZDA: line=%q", clean))
		return
	}
	if len(timeFields[1]) != 9 {
		d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad GPZDA: line=%q", clean))
		return
	}
	timestr := timeFields[1][:2] + ":" + timeFields[1][2:4] + ":" + timeFields[1][4:6]
	datestr := timeFields[4] + "-" + timeFields[3] + "-" + timeFields[2]
	t, err := time.Parse(time.RFC3339, datestr+"T"+timestr+"Z")
	if err != nil {
		d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad GPZDA: %v: line=%q", err, clean))
		return
	}

	// Latitude and Longitude
	latLonFields := strings.Split(fields[2], ",")
	if len(latLonFields) != 15 {
		d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad GPGGA: line=%q", clean))
		return
	} else {
		latdd, latddErr := geo.GGALat2DD(latLonFields[2], latLonFields[3])
		if latddErr != nil {
			d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad GPGGA lat: %v: line=%q", latddErr, line))
			return
		}
		values["lat"] = latdd

		// Longitude
		londd, londdErr := geo.GGALon2DD(latLonFields[4], latLonFields[5])
		if londdErr != nil {
			d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad GPGGA lon: %v: line=%q", londdErr, line))
			return
		}
		values["lon"] = londd
	}

	// // Temperature
	tsgFields := strings.Split(fields[3], ",")
	if len(tsgFields) != 3 && len(tsgFields) != 4 {
		d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad TSG: line=%q", clean))
		values["temp"] = tsdata.NA
		values["conductivity"] = tsdata.NA
		values["salinity"] = tsdata.NA
	} else {
		tempStr := strings.TrimSpace(tsgFields[0])
		_, floatErr := strconv.ParseFloat(strings.TrimSpace(tsgFields[0]), 64)
		if floatErr != nil {
			d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad float: line=%q", line))
			values["temp"] = tsdata.NA
		} else {
			values["temp"] = tempStr
		}

		// Conductivity
		condStr := strings.TrimSpace(tsgFields[1])
		_, floatErr = strconv.ParseFloat(condStr, 64)
		if floatErr != nil {
			d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad float: line=%q", line))
			values["conductivity"] = tsdata.NA
		} else {
			values["conductivity"] = condStr
		}

		// Salinity
		salStr := strings.TrimSpace(tsgFields[2])
		_, floatErr = strconv.ParseFloat(salStr, 64)
		if floatErr != nil {
			d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad float: line=%q", line))
			values["salinity"] = tsdata.NA
		} else {
			values["salinity"] = salStr
		}
	}

	// PAR
	// Keep the line if PAR is simply not present, PAR feed may have stopped
	// entirely and we should keep all other values if possible. As opposed to
	// PAR being present but with possibly truncated numbers, in which case we
	// want to reject the entire line in the hopes that a valid PAR value shows
	// up within the throttled time interval. A full PAR number has 3 decimal
	// places.
	parField := fields[4]
	if parField == "" {
		d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad PPAR: line=%q", clean))
		values["par"] = tsdata.NA
	} else {
		parStr := strings.TrimSpace(parField)
		parNumberFields := strings.Split(parField, ".")
		_, floatErr := strconv.ParseFloat(parStr, 64)
		if floatErr != nil || len(parNumberFields) != 2 || len(parNumberFields[1]) != 3 {
			d.Errors = append(d.Errors, fmt.Errorf("TN427Parser: bad PAR float: line=%q", line))
			values["par"] = tsdata.NA
			// PAR may be unreliable on TN427+. We'll be reading every second, so just completely
			// reject the entire line if bad PAR. On G5 about 1 in 4 PAR was good, so if
			// we encounter the same issue on TN427+ we'll be ready.
			return
		} else {
			values["par"] = parStr
		}
	}

	// Populate Data
	if len(values) == 6 {
		// Prepare complete Data struct
		d.Time = t
		d.Values = make([]string, len(values))
		for i, k := range p.metadata.Headers {
			if k != "time" {
				d.Values[i-1] = values[k]
			}
		}
		p.Limit(&d)
	}

	return
}

// Header returns a string header for a TSDATA file.
func (p *TN427Parser) Header() string {
	return p.metadata.Header()
}
