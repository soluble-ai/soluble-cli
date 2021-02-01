package fingerprint

import (
	"bufio"
	"os"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestPartialFingerprints(t *testing.T) {
	assert := assert.New(t)
	f, err := os.Open("fingerprint.go")
	util.Must(err)
	defer f.Close()
	r := bufio.NewReader(f)
	fingerprints := map[int]string{}
	err = Partial(r, func(n int, fingerprint string) {
		fingerprints[n] = fingerprint
	})
	assert.Nil(err)
	// check that we got a fingerprint for every line
	assert.Greater(len(fingerprints), 50)
	for i := 1; i <= len(fingerprints); i++ {
		fingerprint := fingerprints[i]
		assert.NotEmpty(fingerprint)
		assert.Greater(len(fingerprint), 15)
	}
}
