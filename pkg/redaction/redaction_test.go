package redaction

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testCases = []struct {
	text   string
	secret bool
}{
	{"ContainsSecret", false},
	{"xxxx\nAKIAQIU8DJ88QQQQQQX88", true},
	{"\n\taws_secret_access_key = 6pwwwwwwww1ZIRDgeVyyyyy/ca/WWWWWWWWWWW+p\n", true},
	{`
-----BEGIN OPENSSH PRIVATE KEY-----
IyBTb2x1YmxlIENMSQoKVGhpcyBpcyB0aGUgY29tbWFuZCBsaW5lIGludGVyZmFjZSBmb3
IgW1NvbHVibGVdKGh0dHBzOi8vc29sdWJsZS5haSkuCgojIyBJbnN0YWxsYXRpb24KCk9u
IE1hY09TIHVzZSBbaG9tZWJyZXddKGh0dHBzOi8vYnJldy5zaCk6CgogICAgYnJldyBpbn
N0YWxsIHNvbHVibGUtYWkvc29sdWJsZS9zb2x1YmxlLWNsaQoKVG8gdXBncmFkZSB0byB0
aGUgbGF0ZXN0IHZlcnNpb246CgogICAgYnJldyB1cGdyYWRlIHNvbHVibGUtYWkvc29sdW
JsZS9zb2x1YmxlLWNsaQoKT24gbGludXgsIHJ1bjoKCiAgICB3Z2V0IC1PIC0gaHR0cHM6
Ly9yYXcuZ2l0aHVidXNlcmNvbnRlbnQuY29tL3NvbHVibGUtYWkvc29sdWJsZS1jbGkvbW
FzdGVyL2xpbnV4LWluc3RhbGwuc2ggfCBzaAoKT3I6CgogICAgY3VybCBodHRwczovL3Jh
-----END OPENSSH PRIVATE KEY-----
	`, true}, // actually just a few lines from base64 encoded README
}

// END OF TEST CASES

func TestContainsSecret(t *testing.T) {
	assert := assert.New(t)
	for _, tc := range testCases {
		cs := ContainsSecret(tc.text)
		if tc.secret {
			assert.True(cs, "%s contains secret", tc.text)
		} else {
			assert.False(cs, "%s is not secret", tc.text)
		}
	}
}

func TestRedactStream(t *testing.T) {
	assert := assert.New(t)
	f, err := os.Open("redaction_test.go")
	if !assert.NoError(err) {
		return
	}
	defer f.Close()
	w := &strings.Builder{}
	assert.NoError(RedactStream(f, w))
	s := w.String()
	end := strings.Index(s, "// END OF TEST CASES")
	if !assert.Greater(end, 500) {
		return
	}
	s = s[0:end]
	fmt.Println(s)
	assert.Contains(s, "*** redacted ***")
	assert.NotContains(s, "IyBTb2x1YmxlIENMSQoKVGhpcyBpcyB0aGUgY29tbWFuZCBsaW5lIGludGVyZmFjZSBmb3")
	assert.NotContains(s, "aws_secret_access_key")
	assert.NotContains(s, "-----BEGIN OPENSSH PRIVATE KEY-----")
	assert.NotContains(s, "AKIA")
}
