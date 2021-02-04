package assessments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFailureThresholds(t *testing.T) {
	var testCases = []struct {
		input                             []string
		fail                              bool
		info, low, medium, high, critical int
	}{
		{[]string{"medium=5", "critical=1"}, false, -1, -1, 5, 5, 1},
		{nil, false, -1, -1, -1, -1, -1},
		{[]string{"medium"}, false, -1, -1, 1, 1, 1},
		{[]string{"=1"}, true, -1, -1, -1, -1, -1},
		{[]string{"HIGH=1"}, false, -1, -1, -1, 1, 1},
		{[]string{"HIGH=1", "high=5"}, false, -1, -1, -1, 5, 5},
		{[]string{"hig=1"}, true, -1, -1, -1, -1, -1},
	}
	assert := assert.New(t)
	for _, tc := range testCases {
		m, err := ParseFailThresholds(tc.input)
		if tc.fail {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
			assert.Equal(tc.info, m["info"])
			assert.Equal(tc.low, m["low"])
			assert.Equal(tc.medium, m["medium"])
			assert.Equal(tc.high, m["high"])
			assert.Equal(tc.critical, m["critical"])
		}
	}
}
