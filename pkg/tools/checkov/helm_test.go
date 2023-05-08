package checkov

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	assert := assert.New(t)
	tool := Helm{
		Include:  []string{"**"},
		Parallel: -2,
	}
	assert.NoError(tool.Validate())
	assert.Equal(len(tool.charts), 3)
	charts := []string{"testdata/helm-charts/charts/subchart/Chart.yaml", "testdata/helm-charts/charts/subchart2/Chart.yaml", "testdata/mychart/Chart.yaml"}
	assert.ElementsMatch(tool.charts, charts)
}
