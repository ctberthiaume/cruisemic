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
	ParseLine(line string) error
	Headers() map[string]string
	Flush() Data
	Limit(d *Data)
	Recent(feed string) time.Time
	GeoThermDefString(feedPrefix string, feedSuffix string) (string, error)
}

// ParserRegistry allows underway parser constructors to be retrieved by name.
var ParserRegistry = map[string]func(string, time.Duration) Parser{
	"Gradients4": NewGradients4Parser,
}

// RegistryChoices returns keys for ParserRegistry one per line.
func RegistryChoices() string {
	var choices []string
	for k := range ParserRegistry {
		choices = append(choices, k)
	}
	return strings.Join(choices, "\n")
}
