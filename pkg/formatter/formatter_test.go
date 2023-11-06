package formatter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatTimeWithUTCAndLocal(t *testing.T) {
	assert := assert.New(t)
	tests := map[string]struct {
		input time.Time
		want  string
	}{
		"simple": {input: time.Date(2021, 8, 15, 14, 30, 45, 100, time.UTC), want: "2021-08-15 14:30 UTC (07:30 MST)"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {

			// taking a location without daylight savings
			Location = "US/Arizona"
			res := FormatTimeWithUTCAndLocal(tc.input)
			assert.Equal(tc.want, res)
		})
	}
}
