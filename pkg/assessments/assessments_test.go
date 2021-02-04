package assessments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssessmentHasFailures(t *testing.T) {
	assert := assert.New(t)
	var testCases = []struct {
		findings   []*Finding
		thresholds []string
		fail       bool
		level      string
		count      int
	}{
		{[]*Finding{}, nil, false, "", 0},
		{[]*Finding{{Severity: "high", Pass: true}}, []string{"high=1"}, false, "", 0},
		{[]*Finding{{Severity: "high", Pass: false}}, []string{"low=1"}, true, "high", 1},
		{[]*Finding{{Severity: "high", Pass: false}}, []string{"medium=1"}, true, "high", 1},
	}
	for _, tc := range testCases {
		assessment := &Assessment{
			Findings: tc.findings,
		}
		thresholds, err := ParseFailThresholds(tc.thresholds)
		assert.Nil(err)
		assessment.EvaluateFailures(thresholds)
		assert.Equal(tc.fail, assessment.Failed, tc)
		assert.Equal(tc.level, assessment.FailedSeverity, tc)
		assert.Equal(tc.count, assessment.FailedCount, tc)
	}
}
