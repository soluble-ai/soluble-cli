package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocker(t *testing.T) {
	if hasDocker() == nil {
		assert := assert.New(t)
		dt := &DockerTool{
			Image: "hello-world",
		}
		d, err := dt.run(true)
		assert.Nil(err)
		assert.Contains(string(d), "Hello from Docker!")
	}
}
