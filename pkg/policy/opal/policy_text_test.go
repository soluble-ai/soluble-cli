package opal

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/stretchr/testify/assert"
)

func TestReadWithMetadoc(t *testing.T) {
	assert := assert.New(t)
	r, err := readPolicyText("testdata/policy-with-metadoc.rego")
	assert.NoError(err)
	if !assert.NotNil(r) {
		return
	}
	assert.Equal(11, r.packageDecl.start)
	assert.Equal(33, r.packageDecl.end)
	assert.Equal("package policies.p1.p2", string(r.text[r.packageDecl.start:r.packageDecl.end]))
	assert.Equal(59, r.regoMetaDoc.start)
	assert.Equal(284, r.regoMetaDoc.end)
	regoMetaDoc := string(r.text[r.regoMetaDoc.start:r.regoMetaDoc.end])
	assert.True(strings.HasPrefix(regoMetaDoc, "__rego__metadoc__ :="), regoMetaDoc)
	s := &strings.Builder{}
	assert.NoError(r.write(s, policy.Metadata{
		"sid":      "c-opl-test-policy",
		"severity": "High",
	}))
	dat, err := os.ReadFile("testdata/policy-with-metadoc-rewrite.rego")
	assert.NoError(err)
	assert.Equal(string(dat), s.String())
}

func TestNoMetadoc(t *testing.T) {
	assert := assert.New(t)
	r, err := readPolicyText("testdata/policy-no-metadoc.rego")
	assert.NoError(err)
	if !assert.NotNil(r) {
		return
	}
	assert.Equal("tf", r.inputType)
	s := &strings.Builder{}
	assert.NoError(r.write(s, policy.Metadata{
		"sid":         "c-opl-test-policy",
		"title":       `This is a "great" example`,
		"description": `This is a great "description"`,
	}))
	fmt.Println(s.String())
	dat, err := os.ReadFile("testdata/policy-no-metadoc-rewrite.rego")
	assert.NoError(err)
	assert.Equal(string(dat), s.String())
}

func TestQuote(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(`"hello"`, regoQuote("hello"))
}

func TestReadNoPackage(t *testing.T) {
	assert := assert.New(t)
	r, err := readPolicyText("testdata/policy-no-package.rego")
	assert.NoError(err)
	if !assert.NotNil(r) {
		return
	}
	b := &strings.Builder{}
	assert.NoError(r.write(b, policy.Metadata{
		"sid": "c-opl-test-no-package",
	}))
	s := b.String()
	fmt.Println(s)
	dat, err := os.ReadFile("testdata/policy-no-package-rewrite.rego")
	assert.NoError(err)
	assert.Equal(string(dat), s)
}
