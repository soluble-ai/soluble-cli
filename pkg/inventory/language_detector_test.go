package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguageDetectors(t *testing.T) {
	assert := assert.New(t)
	m := &Manifest{}
	m.scan("testdata/lang", pythonDetector())
	assert.Contains(m.PythonDirectories.Values(), "python-app")
	m = &Manifest{}
	m.scan("testdata/lang/python-app", pythonDetector())
	assert.Contains(m.PythonDirectories.Values(), ".")
	m = &Manifest{}
	m.scan("testdata/lang", javaDetector())
	assert.ElementsMatch(m.JavaDirectories.Values(), []string{
		"java", "java2", "java3",
	})
	m = &Manifest{}
	m.scan("testdata/lang/java3", javaDetector())
	assert.ElementsMatch(m.JavaDirectories.Values(), []string{"."})
}
