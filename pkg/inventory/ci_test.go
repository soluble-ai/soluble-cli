package inventory

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCI(t *testing.T) {
	assert := assert.New(t)
	m := &Manifest{}
	m.scan(filepath.Join("testdata", "ci"), cidetector(0))
	assert.ElementsMatch(m.CISystems.Values(), []string{
		"github", "drone", "gitlab", "circleci", "jenkins", "travis", "azure",
	})
	m.CISystems.Reset()
	m.scan("testdata", cidetector(0))
	if m.CISystems.Len() != 0 {
		t.Error(m.CISystems.Values())
	}
}
