package manager

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
	pt.PolicyDir += "/" + pt.PolicyType + "/" + pt.PolicyName + "/" + pt.CheckType + "/tests"
	if err := os.MkdirAll(pt.PolicyDir, os.ModePerm); err != nil {
		return err
	} else {
		fmt.Println("created: ", pt.PolicyDir)
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
		Category:    pt.PolicyCategory,
		Description: doubleQuote(pt.PolicyDesc),
		Severity:    pt.PolicySeverity,
		Title:       doubleQuote(pt.PolicyTitle),
	}

	data, err := yaml.Marshal(&metadata)

	if err != nil {
		log.Fatal(err)
	}

	metadataPath := strings.Split(pt.PolicyDir, pt.CheckType)[0] + "/metadata.yaml"
	err2 := os.WriteFile(metadataPath, data, 0600)

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
	regoPath := strings.Split(pt.PolicyDir, "tests")[0] + "/policy.rego"
	regoTemplate :=
		"package policies." + pt.PolicyName +
			"\n\n" +
			"input_type := \"" + inputTypeForTarget[policy.Target(pt.CheckType)] + "\""

	if pt.PolicyRsrcType != "" {
		regoTemplate += "\n\nresource_type = " + pt.PolicyRsrcType
	}

	regoTemplate +=
		"\n\ndefault allow = false" +
			"\n\n# Add Rego Policy # \n"

	err := os.WriteFile(regoPath, []byte(regoTemplate), 0600)

	if err != nil {
		log.Fatal(err)
	}
	return nil
}
