// Package parse implements parsers, formatters, and data descriptions for
// research cruise underway data feeds.
package parse

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/storage"
)

// rawName is the string designator for unparsed text data sent to storage
const RawName = "raw"

// underwayName is the string designator for parsed underway text data sent to storage
const UnderwayName = "geo"

// Parser is the interface that groups the ParseLine and RateLimit used to
// parser a ship's underway feed.
type Parser interface {
	ParseLine(line string) Data
	Header() string
	Limit(d *Data)
}

// ParseLines parses cruise feed lines and saves data to storage
func ParseLines(parser Parser, r io.Reader, storer storage.Storer, rawFlag bool, flushFlag bool, noCleanFlag bool) (err error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b := scanner.Bytes()
		n := len(b)
		if !noCleanFlag {
			// Remove unwanted ASCII characters
			n = Whitelist(b, n)
		}
		line := string(b[:n])

		if rawFlag {
			// Save raw text for each line
			err = storer.WriteString(RawName, line+"\n")
			if err != nil {
				return fmt.Errorf("error writing unparsed text: %v", err)
			}
		}

		d := parser.ParseLine(line)
		for _, err := range d.Errors {
			log.Printf("%v", err)
		}
		if d.OK() {
			// Save data if properly parsed and not throttled
			err = storer.WriteString(UnderwayName, d.Line("\t")+"\n")
			if err != nil {
				return fmt.Errorf("error writing parsed data: %v", err)
			}
		}

		if flushFlag {
			err = storer.Flush()
			if err != nil {
				return fmt.Errorf("error flushing data: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading lines: %v", err)
	}
	return nil
}

// ParserRegistry allows underway parser constructors to be retrieved by name.
var ParserRegistry = map[string]func(string, time.Duration, func() time.Time) Parser{
	"Gradients4": NewGradients4Parser,
	"Kilo Moana": NewKiloMoanaParser,
}

// RegistryChoices returns keys for ParserRegistry one per line.
func RegistryChoices() string {
	var choices []string
	for k := range ParserRegistry {
		choices = append(choices, k)
	}
	return strings.Join(choices, "\n")
}
