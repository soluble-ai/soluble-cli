package assessments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssessmentHasFailures(t *testing.T) {
	assert := assert.New(t)
	var testCases = []struct {
		findings   []Finding
		thresholds map[string]string
		fail       bool
		level      string
		count      int
	}{
		{[]Finding{}, map[string]string{}, false, "", 0},
		{[]Finding{{Severity: "high", Pass: false}}, map[string]string{"low": "1"}, true, "high", 1},
	}
	for _, tc := range testCases {
		assessment := &Assessment{
			Findings: tc.findings,
		}
		thresholds, err := ParseFailThresholds(tc.thresholds)
		assert.Nil(err)
		f, level, count := assessment.HasFailures(thresholds)
		assert.Equal(tc.fail, f, tc)
		assert.Equal(tc.level, level, tc)
		assert.Equal(tc.count, count, tc)
	}
}
