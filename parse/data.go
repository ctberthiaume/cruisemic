package parse

import (
	"strings"
	"time"
)

// Data holds parsed data for a single time point of an underway feed
type Data struct {
	Time      time.Time
	Throttled bool
	Values    []string
	Errors    []error
}

func (d Data) String() (s string) {
	if d.Throttled {
		s += "(T)"
	}
	s += "::"
	if !d.Time.IsZero() {
		s += d.Time.Format(time.RFC3339Nano)
	}
	if d.Values != nil {
		s = s + "," + strings.Join(d.Values, ",")
	}
	if d.Errors != nil {
		errorStrings := []string{}
		for _, e := range d.Errors {
			errorStrings = append(errorStrings, e.Error())
		}
		s = s + "::" + strings.Join(errorStrings, ",")
	}
	return s
}

// Line creates a delimited line of text, starting with RFC3339 timestamp.
func (d Data) Line(sep string) string {
	s := append([]string{d.Time.Format(time.RFC3339Nano)}, d.Values...)
	return strings.Join(s, sep)
}

// OK indicates whether this Data is ready to write.
func (d Data) OK() bool {
	return !d.Time.IsZero() && (len(d.Values) > 0) && !d.Throttled
}
