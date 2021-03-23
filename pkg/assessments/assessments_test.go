// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
