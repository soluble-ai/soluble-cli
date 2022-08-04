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
	r, err := readRuleText("testdata/rule-with-metadoc.rego")
	assert.NoError(err)
	if !assert.NotNil(r) {
		return
	}
	assert.Equal(11, r.packageDecl.start)
	assert.Equal(30, r.packageDecl.end)
	assert.Equal("package rules.p1.p2", string(r.text[r.packageDecl.start:r.packageDecl.end]))
	assert.Equal(56, r.regoMetaDoc.start)
	assert.Equal(281, r.regoMetaDoc.end)
	regoMetaDoc := string(r.text[r.regoMetaDoc.start:r.regoMetaDoc.end])
	assert.True(strings.HasPrefix(regoMetaDoc, "__rego__metadoc__ :="), regoMetaDoc)
	s := &strings.Builder{}
	assert.NoError(r.write(s, policy.Metadata{
		"sid":      "c-opl-test-rule",
		"severity": "High",
	}))
	dat, err := os.ReadFile("testdata/rule-with-metadoc-rewrite.rego")
	assert.NoError(err)
	assert.Equal(string(dat), s.String())
}

func TestNoMetadoc(t *testing.T) {
	assert := assert.New(t)
	r, err := readRuleText("testdata/rule-no-metadoc.rego")
	assert.NoError(err)
	if !assert.NotNil(r) {
		return
	}
	assert.Equal("tf", r.inputType)
	s := &strings.Builder{}
	assert.NoError(r.write(s, policy.Metadata{
		"sid":         "c-opl-test-rule",
		"title":       `This is a "great" example`,
		"description": `This is a great "description"`,
	}))
	fmt.Println(s.String())
	dat, err := os.ReadFile("testdata/rule-no-metadoc-rewrite.rego")
	assert.NoError(err)
	assert.Equal(string(dat), s.String())
}

func TestQuote(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(`"hello"`, regoQuote("hello"))
}
