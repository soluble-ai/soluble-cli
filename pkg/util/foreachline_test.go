package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForEachLine(t *testing.T) {
	assert := assert.New(t)
	var funcLine string
	err := ForEachLine("foreachline.go", func(line string) bool {
		if strings.HasPrefix(line, "func ForEachLine") {
			funcLine = line
		}
		return true
	})
	assert.NoError(err)
	assert.NotEmpty(funcLine)
}
