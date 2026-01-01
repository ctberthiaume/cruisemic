package rawudp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// RawUDPReader reads raw UDP payloads wrapped with RAWUDP headers.
type RawUDPReader struct {
	scanner *bufio.Scanner
	buffer  bytes.Buffer
	eof     bool
}

// TimeSource is an interface that provides the current time.
type TimeSource interface {
	Now() time.Time
}

// RealTime implements TimeSource to provide current time .
type RealTime struct{}

// Now returns the current time.
func (r RealTime) Now() time.Time {
	return time.Now()
}

// NewRawUDPReader returns a pointer to a RawUDPReader that reads from r.
func NewRawUDPReader(r io.Reader) *RawUDPReader {
	scanner := bufio.NewScanner(r)
	scanner.Split(scanRawUDP)
	return &RawUDPReader{
		scanner: scanner,
	}
}

// Read reads data into p from the raw UDP payloads.
func (r *RawUDPReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Satisfy the read with available data
	if r.buffer.Len() > 0 {
		return r.buffer.Read(p)
	}

	// No more data to be had from scanner, and buffer should be empty by now.
	if r.eof {
		if r.buffer.Len() != 0 {
			return 0, fmt.Errorf("RawUDPReader: eof but buffer len=%d\n", r.buffer.Len())
		}
		return 0, io.EOF
	}

	// Fill buffer with more payload data and satisfy the read if possible
	for r.scanner.Scan() {
		b := r.scanner.Bytes() // a complete payload
		r.buffer.Write(b)
		// Satisfy the read if possible
		if r.buffer.Len() >= len(p) {
			return r.buffer.Read(p)
		}
	}

	// Either we ran out of payload data before satisfying the read, or an
	// error occurred.
	r.eof = true // scanner is exhausted
	if err := r.scanner.Err(); err != nil {
		return 0, err
	}
	// No more payloads to read, read whatever is left in buffer.
	// If buffer is empty, this will return io.EOF.
	n, err = r.buffer.Read(p)
	return
}

// scanRawUDP is a split function for a Scanner that returns each raw UDP payload
// wrapped with a RAWUDP header.
func scanRawUDP(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if strings.HasPrefix(string(data), "=== RAWUDP,") {
		// Find payload length at end of line, read payload
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			line := string(data[:i])
			parts := strings.Split(line, ",")
			if len(parts) != 3 {
				return 0, nil, fmt.Errorf("bad RAWUDP header")
			}
			payloadLen, err := strconv.Atoi(parts[2])
			if err != nil {
				return 0, nil, fmt.Errorf("bad RAWUDP length")
			}
			// header + \n + payload + final \n to terminate payload block
			totalLen := (i + 1) + payloadLen + 1
			if len(data) >= totalLen {
				// We have the full payload. Return payload, making sure to
				// skip header line and final \n. This reconstructs the original
				// UDP payload.
				if payloadLen == 0 {
					return totalLen, []byte{}, nil
				}
				return totalLen, data[i+1 : totalLen-1], nil
			} else {
				// Don't have full payload, request more data if not EOF, else
				// return an error.
				if !atEOF {
					return 0, nil, nil
				}
				return 0, nil, fmt.Errorf("incomplete RAWUDP payload")
			}
		} else {
			// Don't have full header line, request more data
			return 0, nil, nil
		}
	} else {
		// Can't find start of RAWUDP header, check that was we have matches
		// the start of a RAWUDP header.
		expected := "=== RAWUDP,"
		if len(data) < len(expected) {
			if string(data) == expected[:len(data)] {
				// Partial match, request more data
				return 0, nil, nil
			}
		}
		// Something went wrong, return an error
		return 0, nil, fmt.Errorf("bad RAWUDP start: %v", string(data))
	}
}

// WrapUDPPayload wraps a UDP payload with a RAWUDP header using the provided
// TimeSource to get the current time.
func WrapUDPPayload(ts TimeSource, payload []byte) []byte {
	header := fmt.Sprintf("=== RAWUDP,%s,%d\n", ts.Now().UTC().Format(time.RFC3339), len(payload))
	return []byte(header + string(payload) + "\n")
}
