package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testGGAData struct {
	name        string
	coord       string
	nsew        string
	expected    string
	expectError bool
}

func TestGGALat2DD(t *testing.T) {
	testData := []testGGAData{
		{"2116.6922,n", "2116.6922", "n", "21.2782", false},
		{"2116.6922,s", "2116.6922", "s", "-21.2782", false},
		{"2116.6922,N", "2116.6922", "N", "21.2782", false},
		{"2116.6922,S", "2116.6922", "S", "-21.2782", false},
		{"211,n", "211", "n", "", true},
		{"9116.6922,n", "9116.6922", "n", "", true},
		{"2176.6922,n", "2176.6922", "n", "", true},
		{"2116.6922,K", "2116.6922", "K", "", true},
		{"2a16.6922,K", "2a16.6922", "K", "", true},
		{"2116.a922,K", "2116.a922", "K", "", true},
		{"-9116.6922,n", "-9116.6922", "n", "", true},
	}
	for _, tt := range testData {
		t.Run(tt.name, createGGALat2DDTest(t, tt))
	}
}

func TestGGALon2DD(t *testing.T) {
	testData := []testGGAData{
		{"15752.6526,e", "15752.6526", "e", "157.8775", false},
		{"15752.6526,w", "15752.6526", "w", "-157.8775", false},
		{"15752.6526,E", "15752.6526", "E", "157.8775", false},
		{"15752.6526,W", "15752.6526", "W", "-157.8775", false},
		{"1575,e", "1575", "e", "", true},
		{"18752.6526,e", "18752.6526", "e", "", true},
		{"15762.6526,e", "15762.6526", "e", "", true},
		{"15752.6526,K", "15752.6526", "K", "", true},
		{"1a752.6526,e", "1a752.6526", "e", "", true},
		{"15752.6526,e", "15752.a526", "e", "", true},
		{"-18752.6526,e", "-18752.6526", "e", "", true},
	}
	for _, tt := range testData {
		t.Run(tt.name, createGGALon2DDTest(t, tt))
	}
}

func createGGALat2DDTest(t *testing.T, tt testGGAData) func(*testing.T) {
	assert := assert.New(t)
	return func(t *testing.T) {
		actual, err := GGALat2DD(tt.coord, tt.nsew)
		if tt.expectError {
			assert.NotNil(err, tt.name)
		}
		assert.Equal(tt.expected, actual, tt.name)
	}
}

func createGGALon2DDTest(t *testing.T, tt testGGAData) func(*testing.T) {
	assert := assert.New(t)
	return func(t *testing.T) {
		actual, err := GGALon2DD(tt.coord, tt.nsew)
		if tt.expectError {
			assert.NotNil(err, tt.name)
		}
		assert.Equal(tt.expected, actual, tt.name)
	}
}

type testDecDegData struct {
	name        string
	coord       string
	expectError bool
}

func TestCheckLon(t *testing.T) {
	testData := []testDecDegData{
		{"100.0", "100.0", false},
		{"-100.0", "-100.0", false},
		{"a", "a", true},
		{"190.0", "190.0", true},
		{"-190.0", "-190.0", true},
	}

	for _, tt := range testData {
		t.Run(tt.name, createCheckLonTest(t, tt))
	}
}

func TestCheckLat(t *testing.T) {
	testData := []testDecDegData{
		{"70.0", "70.0", false},
		{"-70.0", "-70.0", false},
		{"a", "a", true},
		{"95.0", "95.0", true},
		{"-95.0", "-95.0", true},
	}

	for _, tt := range testData {
		t.Run(tt.name, createCheckLatTest(t, tt))
	}
}

func createCheckLonTest(t *testing.T, tt testDecDegData) func(*testing.T) {
	assert := assert.New(t)
	return func(t *testing.T) {
		err := CheckLon(tt.coord)
		if tt.expectError {
			assert.NotNil(err, tt.name)
		} else {
			assert.Nil(err, tt.name)
		}
	}
}

func createCheckLatTest(t *testing.T, tt testDecDegData) func(*testing.T) {
	assert := assert.New(t)
	return func(t *testing.T) {
		err := CheckLat(tt.coord)
		if tt.expectError {
			assert.NotNil(err, tt.name)
		} else {
			assert.Nil(err, tt.name)
		}
	}
}
