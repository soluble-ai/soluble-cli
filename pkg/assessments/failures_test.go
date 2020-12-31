package assessments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFailureThresholds(t *testing.T) {
	assert := assert.New(t)
	m, err := ParseFailThresholds(map[string]string{})
	assert.Nil(err)
	assert.Len(m, 5)
	m, err = ParseFailThresholds(map[string]string{
		"medium":   "5",
		"critical": "1",
	})
	assert.Nil(err)
	assert.Len(m, 5)
	assert.Equal(map[string]int{"info": -1, "low": -1, "medium": 5, "high": 5, "critical": 1}, m)
	m, err = ParseFailThresholds(map[string]string{
		"nope":     "5",
		"high":     "nope",
		"critical": "1",
	})
	assert.NotNil(err)
	assert.Equal(m["critical"], 1)
}
