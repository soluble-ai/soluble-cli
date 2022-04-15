package capture

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombined(t *testing.T) {
	assert := assert.New(t)
	p := exec.Command("ls", "-l", ".", "does-not-exist")
	out := &bytes.Buffer{}
	p.Stdout = out
	cap := NewCombinedOutputCaptureForProcess(p)
	assert.NoError(p.Start())
	if e := p.Wait(); assert.Error(e) {
		assert.NotEqual(0, e.(*exec.ExitError).ExitCode())
	}
	dat, err := cap.OutputBytes()
	assert.NoError(err)
	fmt.Println("dat:")
	fmt.Println(string(dat))
	assert.NoError(err)
	assert.Contains(string(dat), "does-not-exist")
	assert.NoError(cap.Close())
}
