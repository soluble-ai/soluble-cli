//go:build integration

package integration

import (
	"testing"

	"github.com/soluble-ai/soluble-cli/cmd/test"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestRepoInventory(t *testing.T) {
	assert := assert.New(t)
	tool := test.NewCommand(t, "repo-inventory", "-d", "../../..")
	tool.Must(tool.Run())
	m := tool.YAML()
	if assert.NotNil(m) {
		count := util.GenericGet(m, "file_count")
		assert.Greater(count, 300)
		topLevelModules := util.GenericGet(m, "terraform_top_level_modules")
		assert.Contains(topLevelModules, "cmd/tfscan/integration/testdata/withvars")
		s3BucketCount := util.GenericGet(m, "terraform_local_resource_counts/aws_s3_bucket")
		assert.GreaterOrEqual(s3BucketCount, 2)
	}
}
