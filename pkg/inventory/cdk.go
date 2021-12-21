package inventory

import "github.com/soluble-ai/soluble-cli/pkg/util"

func cdkDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:          func(m *Manifest) *util.StringSet { return &m.CDKDirectories },
		markerFiles:        []string{"cdk.json"},
		collapseNestedDirs: true,
	}
}
