package custompolicybuilder

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"gopkg.in/yaml.v3"
)

var inputTypeForTarget = map[policy.Target]string{
	policy.Terraform:      "tf",
	policy.Cloudformation: "cfn",
	policy.Kubernetes:     "k8s",
	policy.ARM:            "arm",
}

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
	pt.Dir += "/" + pt.Type + "/" + pt.Name + "/" + pt.CheckType + "/tests"
	if err := os.MkdirAll(pt.Dir, os.ModePerm); err != nil {
		return err
	} else {
		fmt.Println("created: ", pt.Dir)
	}
	return nil
}

func (pt *PolicyTemplate) GenerateMetadataYaml() error {
	// metadata.yaml (in PolicyName dir)
	type Metadata struct {
		Category    string    `yaml:"category"`
		Description yaml.Node `yaml:"description"`
		Severity    string    `yaml:"severity"`
		Title       yaml.Node `yaml:"title"`
	}

	metadata := Metadata{
		Category:    pt.Category,
		Description: doubleQuote(pt.Desc),
		Severity:    pt.Severity,
		Title:       doubleQuote(pt.Title),
	}

	data, err := yaml.Marshal(&metadata)

	if err != nil {
		log.Fatal(err)
	}

	metadataPath := strings.Split(pt.Dir, pt.CheckType)[0] + "/metadata.yaml"
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
			"\n\ninput_type := \"" + inputTypeForTarget[policy.Target(pt.CheckType)] + "\""

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
