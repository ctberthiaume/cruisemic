package parse

import (
	"time"

	"github.com/ctberthiaume/tsdata"
)

// DataManager supports adding and retrieving parsed data and metadata.
type DataManager struct {
	Throttle
	t        time.Time         // latest time read
	values   map[string]string // latest values by column name
	errors   []error           // errors encountered when parsing latest values
	metadata tsdata.Tsdata     // TSDATA output file metadata
}

// NewDataManager returns a pointer to a DataManager struct. metadata is the
// Tsdata definition of all data values managed by this struct. interval is the
// per-feed rate limiting interval in seconds.
func NewDataManager(metadata tsdata.Tsdata, interval time.Duration) *DataManager {
	return &DataManager{
		Throttle: NewThrottle(interval),
		values:   make(map[string]string),
		metadata: metadata,
	}
}

// Header returns a Tsdata header paragraph string.
func (dm *DataManager) Header() string {
	return dm.metadata.Header()
}

// AddValue adds a parsed value to the DataManager.
func (dm *DataManager) AddValue(key, value string) {
	dm.values[key] = value
}

func (dm *DataManager) GetValue(key string) (string, bool) {
	val, ok := dm.values[key]
	return val, ok
}

// AddError adds a parsing error to the DataManager.
func (dm *DataManager) AddError(err error) {
	dm.errors = append(dm.errors, err)
}

// SetTime sets the time of the latest parsed data.
func (dm *DataManager) SetTime(t time.Time) {
	dm.t = t
}

// GetData returns a Data struct. It returns an empty Data struct if not all
// expected values are present, as defined by column names in dm.metadata.Headers.
// If the caller would like a populated Data struct returned even if some values
// have not been parsed from the data stream, they should add those values as
// Tsdata.NA or some constant string before calling GetData. GetData will also
// call Throttle.Limit to apply rate limiting to the returned Data. When a fully
// populated Data struct is returned, the DataManager's internal state is reset
// for the next stanza of data.
func (dm *DataManager) GetData() (d Data) {
	allValuesPresent := true

	for _, header := range dm.metadata.Headers {
		if header == "time" {
			if dm.t.IsZero() {
				allValuesPresent = false
				break
			}
		} else {
			_, ok := dm.values[header]
			if !ok {
				allValuesPresent = false
				break
			}
		}
	}
	if allValuesPresent {
		// Prepare complete Data struct
		d.Time = dm.t
		d.Values = make([]string, len(dm.metadata.Headers)-1) // All columns except time
		for i, k := range dm.metadata.Headers {
			if k != "time" {
				d.Values[i-1] = dm.values[k]
			}
		}
		d.Errors = dm.errors
		dm.Limit(&d)
		// Reset state after creating populated Data
		dm.t = time.Time{}
		dm.values = make(map[string]string)
		dm.errors = []error{}
	}
	return
}
