package metrics

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func TestExtractor_First(t *testing.T) {
	tests := []struct {
		family string
		labels []LabelPair
		want   *float64
	}{
		{"not_found", nil, nil},
		{"process_start_time_seconds", []LabelPair{{"", ""}}, nil},
		{"process_start_time_seconds", nil, fp(1.58454130883e+09)},
		{"log_entries_total", []LabelPair{{"prefix", ""}}, nil},
		{"log_entries_total", []LabelPair{{"prefix", "validator"}, {"level", "error"}}, fp(84233)},
	}
	me, err := NewFromFile("../testdata/validator/metrics.txt")
	require.NoError(t, err)

	// run tests
	for _, tt := range tests {
		t.Run(tt.family, func(t *testing.T) {
			got := me.First(tt.family, tt.labels)
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, *got, *tt.want)
			}
		})
	}
}

func fp(f float64) *float64 {
	return &f
}
