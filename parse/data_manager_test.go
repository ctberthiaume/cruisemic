package parse

import (
	"errors"
	"testing"
	"time"

	"github.com/ctberthiaume/tsdata"
	"github.com/stretchr/testify/assert"
)

type testDataManagerData struct {
	name         string
	metadata     tsdata.Tsdata
	interval     time.Duration
	timeVal      time.Time
	values       map[string]string
	errs         []error
	expectedData Data
}

func TestDataManager(t *testing.T) {
	now := time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC)
	testData := []testDataManagerData{
		{
			name: "complete data",
			metadata: tsdata.Tsdata{
				Headers: []string{"time", "lat", "lon"},
			},
			interval: 0,
			timeVal:  now,
			values: map[string]string{
				"lat": "47.5",
				"lon": "-122.3",
			},
			errs: nil,
			expectedData: Data{
				Time:      now,
				Values:    []string{"47.5", "-122.3"},
				Errors:    nil,
				Throttled: false,
			},
		},
		{
			name: "missing time",
			metadata: tsdata.Tsdata{
				Headers: []string{"time", "lat", "lon"},
			},
			interval: 0,
			timeVal:  time.Time{},
			values: map[string]string{
				"lat": "47.5",
				"lon": "-122.3",
			},
			errs:         nil,
			expectedData: Data{}, // Empty data expected
		},
		{
			name: "missing value",
			metadata: tsdata.Tsdata{
				Headers: []string{"time", "lat", "lon"},
			},
			interval: 0,
			timeVal:  now,
			values: map[string]string{
				"lat": "47.5",
			},
			errs:         nil,
			expectedData: Data{}, // Empty data expected
		},
		{
			name: "with errors",
			metadata: tsdata.Tsdata{
				Headers: []string{"time", "lat", "lon"},
			},
			interval: 0,
			timeVal:  now,
			values: map[string]string{
				"lat": "47.5",
				"lon": "-122.3",
			},
			errs: []error{errors.New("error 1"), errors.New("error 2")},
			expectedData: Data{
				Time:      now,
				Values:    []string{"47.5", "-122.3"},
				Errors:    []error{errors.New("error 1"), errors.New("error 2")},
				Throttled: false,
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, createDataManagerTest(t, tt))
	}
}

func createDataManagerTest(t *testing.T, tt testDataManagerData) func(*testing.T) {
	assert := assert.New(t)

	return func(t *testing.T) {
		dm := NewDataManager(tt.metadata, tt.interval)

		// Set time
		if !tt.timeVal.IsZero() {
			dm.SetTime(tt.timeVal)
		}

		// Add values
		for k, v := range tt.values {
			dm.AddValue(k, v)
		}

		// Add errors
		for _, err := range tt.errs {
			dm.AddError(err)
		}

		// Check GetValue
		for k, v := range tt.values {
			gotVal, ok := dm.GetValue(k)
			assert.True(ok, "GetValue should return true for existing key: "+k)
			assert.Equal(v, gotVal, "GetValue should return correct value for key: "+k)
		}
		_, ok := dm.GetValue("nonexistent")
		assert.False(ok, "GetValue should return false for nonexistent key")

		// Check Header
		assert.Equal(tt.metadata.Header(), dm.Header(), "Header should match metadata header")

		// Check GetData
		d := dm.GetData()
		assert.Equal(tt.expectedData.Time, d.Time, "Data time mismatch")
		assert.Equal(tt.expectedData.Values, d.Values, "Data values mismatch")
		assert.Equal(tt.expectedData.Throttled, d.Throttled, "Data throttled mismatch")
		assert.Equal(len(tt.expectedData.Errors), len(d.Errors), "Data errors count mismatch")
		for i, err := range tt.expectedData.Errors {
			assert.Equal(err.Error(), d.Errors[i].Error(), "Data error mismatch")
		}

		// Check reset state if data was returned
		if d.OK() {
			d2 := dm.GetData()
			assert.True(d2.Time.IsZero(), "DataManager should be reset after successful GetData")
			assert.Empty(d2.Values, "DataManager values should be empty after reset")
			assert.Empty(d2.Errors, "DataManager errors should be empty after reset")
			assert.Empty(dm.values, "DataManager internal values map should be empty after reset")
			assert.Empty(dm.errors, "DataManager internal errors slice should be empty after reset")
			assert.True(dm.t.IsZero(), "DataManager internal time should be zero after reset")
		}
	}
}
