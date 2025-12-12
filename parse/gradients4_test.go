package parse

import (
	"strings"
	"testing"
	"time"

	"github.com/ctberthiaume/cruisemic/storage"
	"github.com/stretchr/testify/assert"
)

type testG4LineData struct {
	name     string
	input    string
	expected map[string][]string
}

var g4TimeStartStr = "2022-05-27T00:00:00+00:00"

func TestG4ParserRegistry(t *testing.T) {
	assert := assert.New(t)
	constructor, ok := ParserRegistry["Gradients4"]
	assert.True(ok, "Gradients4 parser is registered")
	if ok {
		p := constructor("testproject", 0, time.Now)
		_, ok = p.(*Gradients4Parser)
		assert.True(ok, "Gradients4 parser is registered")
	}
}

func TestG4Lines(t *testing.T) {
	t0, err := time.Parse(time.RFC3339, g4TimeStartStr)
	if err != nil {
		panic(err)
	}
	testData := []testG4LineData{
		{
			"good stanza",
			`$SEAFLOW
2118.9043N
15752.6526W
26.8
5.3
30.4
$SEAFLOW
`,
			map[string][]string{
				"geo": {t0.Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\t5.3\t30.4\n"},
			},
		},
		{
			"2 good stanzas",
			`$SEAFLOW
2118.9043N
15752.6526W
26.8
5.3
30.5
$SEAFLOW
2118.9043N
15752.6526W
26.8
5.3
30.6
$SEAFLOW
`,
			map[string][]string{
				"geo": {
					t0.Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\t5.3\t30.5\n",
					(t0.Add(time.Second)).Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\t5.3\t30.6\n",
				},
			},
		},
		{
			"2 good stanzas, with empty lines in between",
			`$SEAFLOW
2118.9043N
15752.6526W
26.8
5.3
30.5


$SEAFLOW
2118.9043N
15752.6526W
26.8
5.3
30.6
$SEAFLOW
`,
			map[string][]string{
				"geo": {
					t0.Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\t5.3\t30.5\n",
					(t0.Add(time.Second)).Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\t5.3\t30.6\n",
				},
			},
		},
		{
			"bad initial line",
			`$SEAFLOWWWWWWWWWWWW
2118.9043N
15752.6526W
26.8
5.3
30.9
$SEAFLOW
`,
			map[string][]string{},
		},
		{
			"empty line in stanza",
			`$SEAFLOW
2118.9043N

15752.6526W
26.8
5.3
30.9
$SEAFLOW
`,
			map[string][]string{},
		},
		{
			"bad lat number",
			`$SEAFLOW
211a8.9043N
15752.6526W
26.8
5.3
30.9
$SEAFLOW
`,
			map[string][]string{},
		},
		{
			"bad lon number",
			`$SEAFLOW
2118.9043N
157a52.6526W
26.8
5.3
30.9
$SEAFLOW
`,
			map[string][]string{},
		},
		{
			"bad lat direction",
			`$SEAFLOW
2118.9043A
15752.6526W
26.8
5.3
30.9
$SEAFLOW
`,
			map[string][]string{},
		},
		{
			"bad lon direction",
			`$SEAFLOW
2118.9043N
15752.6526A
26.8
5.3
30.9
$SEAFLOW
`,
			map[string][]string{},
		},
		{
			"bad temp",
			`$SEAFLOW
2118.9043N
15752.6526W
26a.8
5.3
30.9
$SEAFLOW
`,
			map[string][]string{
				"geo": {t0.Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\tNA\t5.3\t30.9\n"},
			},
		},
		{
			"bad conductivity",
			`$SEAFLOW
2118.9043N
15752.6526W
26.8
5a.3
30.9
$SEAFLOW
`,
			map[string][]string{
				"geo": {t0.Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\tNA\t30.9\n"},
			},
		},
		{
			"bad salinity",
			`$SEAFLOW
2118.9043N
15752.6526W
26.8
5.3
30a.9
$SEAFLOW
`,
			map[string][]string{
				"geo": {t0.Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\t5.3\tNA\n"},
			},
		},
		{
			"incomplete stanza, missing temp line",
			`$SEAFLOW
2118.9043N
15752.6526W
5.3
30.4
$SEAFLOW
`,
			map[string][]string{},
		},
		{
			"incomplete stanza, good stanza",
			`$SEAFLOW
2118.9043N
15752.6526W
5.3
30.4
$SEAFLOW
2118.9043N
15752.6526W
26.8
5.3
30.5
$SEAFLOW
`,
			map[string][]string{
				"geo": {(t0.Add(time.Second)).Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\t5.3\t30.5\n"},
			},
		},
		{
			"good stanza, incomplete stanza",
			`$SEAFLOW
2118.9043N
15752.6526W
26.8
5.3
30.4
$SEAFLOW
2118.9043N
15752.6526W
5.3
30.5
$SEAFLOW
`,
			map[string][]string{
				"geo": {t0.Format(time.RFC3339Nano) + "\t21.3151\t-157.8775\t26.8\t5.3\t30.4\n"},
			},
		},
	}
	for _, tt := range testData {
		t.Run(tt.name, createG4LinesTest(t, tt))
	}
}

func createG4LinesTest(t *testing.T, tt testG4LineData) func(*testing.T) {
	assert := assert.New(t)

	t0, err := time.Parse(time.RFC3339, g4TimeStartStr)
	if err != nil {
		panic(err)
	}

	i := 0
	now := func() time.Time {
		ti := t0.Add(time.Second * time.Duration(i))
		i++
		return ti
	}

	return func(t *testing.T) {
		p := NewGradients4Parser("test", 0, now)
		store, _ := storage.NewMemStorage()
		r := strings.NewReader(tt.input)
		err := ParseLines(p, r, store, true, false)
		assert.Nil(err, "writing for test: "+tt.name)
		// No need to check the raw feed
		// delete(store.Feeds, "raw")

		assert.Equal(tt.expected, store.Feeds, tt.name)
	}
}
