package tools

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitPRDiffs(t *testing.T) {
	assert := assert.New(t)
	opts := &UploadOpt{
		GitPRBaseRef: "HEAD~1",
	}
	dat := opts.getPRDIffText(".")
	assert.NotEmpty(dat)
	diff := string(dat)
	assert.True(strings.HasPrefix(diff, "# git diff "))
}
