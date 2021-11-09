package parse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// type testG4LineData struct {
// 	name        string
// 	input       string
// 	expected    Data
// 	expectError bool
// }

func TestG4ParserRegistry(t *testing.T) {
	assert := assert.New(t)
	constructor, ok := ParserRegistry["Gradients4"]
	assert.True(ok, "Gradients4 parser is registered")
	if ok {
		p := constructor("testproject", 0)
		_, ok = p.(*Gradients4Parser)
		assert.True(ok, "Gradients4 parser is registered")
	}
}

// func TestG4Lines(t *testing.T) {
// 	testData := []testG4LineData{
// 		{"empty line", "", Data{}, false},
// 		{"bad field count", "40.10, 120.23, 28.1, 4.4, 30.3, 1.04, 20.0, 30.0", Data{}, true},
// 		{"bad lat", "240.10, 120.23, 28.1, 4.4, 30.3, 1.04", Data{}, true},
// 		{"bad lon", "40.10, 420.23, 28.1, 4.4, 30.3, 1.04", Data{}, true},
// 		{"bad temp", "40.10, 120.23, 2a8.1, 4.4, 30.3, 1.04", Data{}, true},
// 		{"bad cond", "40.10, 120.23, 28.1, 4a.4, 30.3, 1.04", Data{}, true},
// 		{"bad sal", "40.10, 120.23, 28.1, 4.4, 3a0.3, 1.04", Data{}, true},
// 		{"bad par", "40.10, 120.23, 28.1, 4.4, 30.3, 1a.04", Data{}, true},
// 		{"good line", "40.10, 120.23, 28.1, 4.4, 30.3, 1.04", Data{Feed: "geo", Values: []string{"40.10", "120.23", "28.1", "4.4", "30.3", "1.04"}}, false},
// 	}
// 	for _, tt := range testData {
// 		t.Run(tt.name, createG4LinesTest(t, tt))
// 	}
// }

// func createG4LinesTest(t *testing.T, tt testG4LineData) func(*testing.T) {
// 	assert := assert.New(t)
// 	return func(t *testing.T) {
// 		p := NewGradients4Parser("test", 0)
// 		actual, err := p.ParseLine(tt.input)
// 		if !tt.expectError {
// 			assert.Nil(err, tt.name)
// 		} else {
// 			assert.NotNil(err, tt.name)
// 		}
// 		assert.Equal(tt.expected.Feed, actual.Feed, tt.name)
// 		assert.Equal(tt.expected.Values, actual.Values, tt.name)
// 		// assert.Equal(tt.expected.Time.Format(time.RFC3339Nano), actual.Time.Format(time.RFC3339Nano), tt.name)
// 	}
// }
