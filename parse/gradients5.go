package parse

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/geo"
	"github.com/ctberthiaume/tsdata"
)

// Gradients5Parser is a parser for Gradients 4 Thompson underway feed lines.
type Gradients5Parser struct {
	Throttle
	metadata tsdata.Tsdata
}

// NewGradients5Parser returns a pointer to a Gradients5Parser struct. project is
// the project or cruise name. interval is the per-feed rate limiting interval
// in seconds.
func NewGradients5Parser(project string, interval time.Duration, now func() time.Time) Parser {
	p := &Gradients5Parser{
		Throttle: NewThrottle(interval),
	}
	p.metadata = tsdata.Tsdata{
		Project:         project,
		FileType:        "geo",
		FileDescription: "Gradients 5 Thompson underway feed",
		Comments:        []string{"RFC3339", "Latitude Decimal format", "Longitude Decimal format", "TSG temperature", "TSG conductivity", "TSG salinity", "PAR"},
		Types:           []string{"time", "float", "float", "float", "float", "float", "float"},
		Units:           []string{"NA", "deg", "deg", "C", "S/m", "PSU", "µE/m^2/s"},
		Headers:         []string{"time", "lat", "lon", "temp", "conductivity", "salinity", "par"},
	}

	return p
}

// ParseLine parses and saves a single underway feed line.
func (p *Gradients5Parser) ParseLine(line string) (d Data) {
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
		d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad GPZDA: line=%q", clean))
		return
	}
	if len(timeFields[1]) != 9 {
		d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad GPZDA: line=%q", clean))
		return
	}
	timestr := timeFields[1][:2] + ":" + timeFields[1][2:4] + ":" + timeFields[1][4:6]
	datestr := timeFields[4] + "-" + timeFields[3] + "-" + timeFields[2]
	t, err := time.Parse(time.RFC3339, datestr+"T"+timestr+"Z")
	if err != nil {
		d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad GPZDA: %v: line=%q", err, clean))
		return
	}

	// Latitude and Longitude
	latLonFields := strings.Split(fields[2], ",")
	if len(latLonFields) != 15 {
		d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad GPGGA: line=%q", clean))
		return
	} else {
		latdd, latddErr := geo.GGALat2DD(latLonFields[2], latLonFields[3])
		if latddErr != nil {
			d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad GPGGA lat: %v: line=%q", latddErr, line))
			return
		}
		values["lat"] = latdd

		// Longitude
		londd, londdErr := geo.GGALon2DD(latLonFields[4], latLonFields[5])
		if londdErr != nil {
			d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad GPGGA lon: %v: line=%q", londdErr, line))
			return
		}
		values["lon"] = londd
	}

	// // Temperature
	tsgFields := strings.Split(fields[3], ",")
	if len(tsgFields) != 4 {
		d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad TSG: line=%q", clean))
		values["temp"] = tsdata.NA
		values["conductivity"] = tsdata.NA
		values["salinity"] = tsdata.NA
	} // else {
	// _, floatErr := strconv.ParseFloat(clean, 64)
	// if floatErr != nil {
	// 	d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad float: line=%q", line))
	// 	values["temp"] = tsdata.NA
	// } else {
	// 	values["temp"] = clean
	// }

	// // Conductivity
	// _, floatErr := strconv.ParseFloat(clean, 64)
	// if floatErr != nil {
	// 	d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad float: line=%q", line))
	// 	values["conductivity"] = tsdata.NA
	// } else {
	// 	values["conductivity"] = clean
	// }

	// // Salinity
	// _, floatErr := strconv.ParseFloat(clean, 64)
	// if floatErr != nil {
	// 	d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad float: line=%q", line))
	// 	values["salinity"] = tsdata.NA
	// } else {
	// 	values["salinity"] = clean
	// }
	// }

	// PAR
	parFields := strings.Split(fields[4], ",")
	if len(parFields) < 2 {
		d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad PPAR: line=%q", clean))
	}
	parStr := strings.TrimSpace(parFields[1])
	parNumberFields := strings.Split(parFields[1], ".")
	_, floatErr := strconv.ParseFloat(parStr, 64)
	if floatErr != nil || len(parNumberFields) != 2 || len(parNumberFields[1]) != 3 {
		d.Errors = append(d.Errors, fmt.Errorf("Gradients5Parser: bad PAR float: line=%q", line))
		values["par"] = tsdata.NA
		// PAR is unreliable on G5. We'll be reading every second, so just completely
		// reject the entire line if bad PAR. About 1 in 4 PAR is good.
		return
	} else {
		values["par"] = parStr
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
func (p *Gradients5Parser) Header() string {
	return p.metadata.Header()
}

// goodPar determines if a PAR value is a valid number with three decimals).
func goodPar(s string) bool {
	fields := strings.Split(s, ".")
	return len(fields) == 2 && len(fields[1]) == 3
}
