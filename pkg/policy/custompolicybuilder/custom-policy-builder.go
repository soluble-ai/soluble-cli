package custompolicybuilder

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"gopkg.in/yaml.v3"
)

func (pt *PolicyTemplate) CreateCustomPolicyTemplate() error {
	if err := pt.CreateDirectoryStructure(); err != nil {
		return err
	}
	if err := pt.GenerateMetadataYaml(); err != nil {
		return err
	}
	if err := pt.GeneratePolicyTemplate(); err != nil {
		return err
	}
	return nil
}

func (pt *PolicyTemplate) CreateDirectoryStructure() error {
	// full directory path
	pt.Dir += "/" + pt.Tool + "/" + pt.Name + "/" + pt.CheckType + "/tests"
	if err := os.MkdirAll(pt.Dir, os.ModePerm); err != nil {
		return err
	} else {
		fmt.Println("created: ", pt.Dir)
	}
	return nil
}

func (pt *PolicyTemplate) GenerateMetadataYaml() error {
	// metadata.yaml (in PolicyName dir)
	metadataPath := strings.Split(pt.Dir, pt.CheckType)[0] + "/metadata.yaml"
	// shouldn't overwrite an existing metadata.yaml file
	if _, err := os.Stat(metadataPath); !os.IsNotExist(err) {
		return nil
	}

	type Metadata struct {
		Category    string    `yaml:"category"`
		CheckTool   string    `yaml:"checkTool"`
		CheckType   string    `yaml:"checkType"`
		Description yaml.Node `yaml:"description"`
		Provider    string    `yaml:"provider"`
		Severity    string    `yaml:"severity"`
		Title       yaml.Node `yaml:"title"`
	}

	metadata := Metadata{
		Category:    pt.Category,
		CheckTool:   pt.Tool,
		CheckType:   pt.CheckType,
		Description: doubleQuote(pt.Desc),
		Provider:    pt.Provider,
		Severity:    pt.Severity,
		Title:       doubleQuote(pt.Title),
	}

	data, err := yaml.Marshal(&metadata)

	if err != nil {
		log.Fatal(err)
	}

	err2 := os.WriteFile(metadataPath, data, os.ModePerm)

	if err2 != nil {
		log.Fatal(err2)
	}
	return nil
}

func doubleQuote(val string) yaml.Node {
	node := yaml.Node{
		Value: val,
		Kind:  yaml.ScalarNode,
		Style: yaml.DoubleQuotedStyle,
	}
	return node
}

func (pt *PolicyTemplate) GeneratePolicyTemplate() error {
	regoPath := strings.Split(pt.Dir, "tests")[0] + "/policy.rego"
	regoTemplate :=
		"package policies." + pt.Name +
			"\n\ninput_type := \"" + policy.InputTypeForTarget[policy.Target(pt.CheckType)] + "\""

	if pt.RsrcType != "" {
		regoTemplate += "\n\nresource_type := \"" + pt.RsrcType + "\""
	}

	regoTemplate +=
		"\n\ndefault allow = false" +
			"\n\n# Add Rego Policy # \n"

	err := os.WriteFile(regoPath, []byte(regoTemplate), os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}
	return nil
}
