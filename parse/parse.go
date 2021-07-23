// Package parse implements parsers, formatters, and data descriptions for
// research cruise underway data feeds.
package parse

import (
	"strings"
	"time"
)

// Parser is the interface that groups the ParseLine and RateLimit used to
// parser a ship's underway feed.
type Parser interface {
	ParseLine(line string) (Data, error)
	Headers() map[string]string
	Limit(d *Data)
	Recent(feed string) time.Time
	GeoThermDefString() string
}

// ParserRegistry allows underway parser constructors to be retrieved by name.
var ParserRegistry = map[string]func(string, time.Duration) Parser{
	"Kilo Moana": NewKiloMoanaParser,
	"Basic":      NewBasicParser,
	"Sally Ride": NewSallyRideParser,
}

// RegistryChoices returns keys for ParserRegistry one per line.
func RegistryChoices() string {
	var choices []string
	for k := range ParserRegistry {
		choices = append(choices, k)
	}
	return strings.Join(choices, "\n")
}
