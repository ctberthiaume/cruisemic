package rawudp

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testRawUDPReaderData struct {
	name      string
	input     string
	expected  string
	expectErr bool
}

type testRawUDPWrapData struct {
	name      string
	input     string
	expected  string
	expectErr bool
}

type fixedTimeSource struct {
	fixedTime time.Time
}

func (fts fixedTimeSource) Now() time.Time {
	return fts.fixedTime
}

func TestRawUDPReader(t *testing.T) {
	testData := []testRawUDPReaderData{
		{
			"good single line UDP block, payload with final \\n",
			"=== RAWUDP,2024-06-01T12:00:00Z,12\nhello world\n\n",
			"hello world\n",
			false,
		},
		{
			"good two line UDP block, payload with final \\n",
			"=== RAWUDP,2024-06-01T12:00:00Z,20\nhello world\ngoodbye\n\n",
			"hello world\ngoodbye\n",
			false,
		},
		{
			"good single line UDP block, payload without final \\n",
			"=== RAWUDP,2024-06-01T12:00:00Z,11\nhello world\n",
			"hello world",
			false,
		},
		{
			"good two line UDP block, payload without final \\n",
			"=== RAWUDP,2024-06-01T12:00:00Z,19\nhello world\ngoodbye\n",
			"hello world\ngoodbye",
			false,
		},
		{
			"empty payload",
			"=== RAWUDP,2024-06-01T12:00:00Z,0\n\n",
			"",
			true, // io.EOF expected on first Read
		},
		{
			"empty payload, then filled payload, then empty payload",
			`=== RAWUDP,2024-06-01T12:00:00Z,0

=== RAWUDP,2024-06-01T12:00:00Z,12
hello world

=== RAWUDP,2024-06-01T12:00:00Z,0

`,
			"hello world\n",
			false,
		},
		{
			"filled payload, then empty payload, then filled payload",
			`=== RAWUDP,2024-06-01T12:00:00Z,12
hello world

=== RAWUDP,2024-06-01T12:00:00Z,0

=== RAWUDP,2024-06-01T12:00:00Z,12
hello world

`,
			"hello world\nhello world\n",
			false,
		},

		{
			"two empty payloads",
			"=== RAWUDP,2024-06-01T12:00:00Z,0\n\n=== RAWUDP,2024-06-01T12:00:00Z,0\n\n",
			"",
			true, // io.EOF expected on first Read
		},
		{
			"payload only contains \\n",
			"=== RAWUDP,2024-06-01T12:00:00Z,1\n\n\n",
			"\n",
			false,
		},
		{
			"single line UDP block, missing final \\n in block",
			"=== RAWUDP,2024-06-01T12:00:00Z,12\nhello world",
			"",
			true,
		},
		{
			"bad UDP header",
			"=== RABAWUDP,2024-06-01T12:00:00Z,12\nhello world",
			"",
			true,
		},
	}
	for _, tt := range testData {
		t.Run(tt.name, createRawUDPReaderTest(t, tt))
	}
}

func createRawUDPReaderTest(t *testing.T, tt testRawUDPReaderData) func(*testing.T) {
	assert := assert.New(t)

	return func(t *testing.T) {
		r := strings.NewReader(tt.input)
		rudpr := NewRawUDPReader(r)
		buf := make([]byte, 1024)
		n, err := rudpr.Read(buf)
		if tt.expectErr {
			assert.NotNil(err, "expected error for test: "+tt.name)
			return
		}
		assert.Nil(err, "reading for test: "+tt.name)
		assert.Equal(len(tt.expected), n, "number of bytes read for test: "+tt.name)
		assert.Equal(tt.expected, string(buf[0:n]), "data read for test: "+tt.name)

		n, err = rudpr.Read(buf)
		assert.Equal(0, n, "expected zero bytes on second read for test: "+tt.name)
		assert.Equal(io.EOF, err, "expected io.EOF on second read for test: "+tt.name)
	}
}

func TestRawUDPWrap(t *testing.T) {
	testData := []testRawUDPWrapData{
		{
			"simple payload",
			"hello world\n",
			"=== RAWUDP,2024-06-01T12:00:00Z,12\nhello world\n\n",
			false,
		},
	}
	for _, tt := range testData {
		t.Run(tt.name, createRawUDPWrapTest(t, tt))
	}
}

func createRawUDPWrapTest(t *testing.T, tt testRawUDPWrapData) func(*testing.T) {
	assert := assert.New(t)

	return func(t *testing.T) {
		fixedTime := time.Date(2024, time.June, 1, 12, 0, 0, 0, time.UTC)
		fts := fixedTimeSource{fixedTime: fixedTime}
		wrapped := WrapUDPPayload(fts, []byte(tt.input))
		assert.Equal(tt.expected, string(wrapped), "data read for test: "+tt.name)
	}
}
