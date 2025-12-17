// Package parse implements parsers, formatters, and data descriptions for
// research cruise underway data feeds.
package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/ctberthiaume/cruisemic/storage"
)

// RawName is the string designator for unparsed text data sent to storage
const RawName = "raw"

// UnderwayName is the string designator for parsed underway text data sent to storage
const UnderwayName = "geo"

// Parser is the interface that groups the ParseLine and RateLimit used to
// parse a ship's underway feed.
type Parser interface {
	ParseLine(line string) Data
	Header() string
	Limit(d *Data)
}

// ParseLines parses cruise feed lines and saves data to storage
func ParseLines(parser Parser, r io.Reader, storer storage.Storer, flushFlag bool, noCleanFlag bool) (err error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(scanLinesWithLF)
	for scanner.Scan() {
		b := scanner.Bytes()
		n := len(b)
		if !noCleanFlag {
			// Remove unwanted ASCII characters
			n = Whitelist(b, n)
		}

		line := string(b[:n])

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

// dropCR drops \r if the last two bytes in data are \r\n.
func dropCR(data []byte) []byte {
	if len(data) > 1 && data[len(data)-2] == '\r' && data[len(data)-1] == '\n' {
		// Replace \r\n with \n
		data[len(data)-2] = data[len(data)-1]
		return data[0 : len(data)-1]
	}
	return data
}

// Like bufio.ScanLines but keeps \n at end of lines to distinguish between
// complete and incomplete lines.
func scanLinesWithLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		// Return it including the \n.
		return i + 1, dropCR(data[0 : i+1]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// ParserRegistry allows underway parser constructors to be retrieved by name.
var ParserRegistry = map[string]func(string, time.Duration, func() time.Time) Parser{
	"Gradients4": NewGradients4Parser,
	"Gradients5": NewGradients5Parser,
	"Kilo Moana": NewKiloMoanaParser,
	"TN427":      NewTN427Parser,
	"TARA":       NewTARAParser,
}

// RegistryChoices returns keys for ParserRegistry one per line.
func RegistryChoices() string {
	var choices []string
	for k := range ParserRegistry {
		choices = append(choices, k)
	}
	return strings.Join(choices, "\n")
}
