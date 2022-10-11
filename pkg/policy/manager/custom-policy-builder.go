package manager

import (
	"fmt"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
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

func (pt *PolicyTemplate) GeneratePolicyTemplate() error {
	regoPath := strings.Split(pt.PolicyDir, "tests")[0] + "/policy.rego"

	regoTemplate := "package policy." + pt.PolicyName +
		"\n\ninput_type = " + inputTypeForTarget[policy.Target(pt.CheckType)]

	if pt.PolicyRsrcType != "" {
		regoTemplate += "\n\nresource_type = " + pt.PolicyRsrcType
	}

	regoTemplate += "\n\ndefault allow = false" +
		"\n\n# Add Rego Policy # \n"

	err := os.WriteFile(regoPath, []byte(regoTemplate), 777)

	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (pt *PolicyTemplate) GenerateMetadataYaml() error {
	// metadata.yaml (in PolicyName dir)
	type Metadata struct {
		Category    string `yaml:"category"`
		Description string `yaml:"description"`
		Severity    string `yaml:"severity"`
		Title       string `yaml:"title"`
		Id          string `yaml:"id"`
	}

	//Check optional flag
	category := pt.PolicyCategory
	if category == "" {
		category = "General"
	}
	severity := pt.PolicySeverity
	if severity == "" {
		severity = "Medium"
	}
	desc := pt.PolicyDesc
	if desc == "" {
		desc = pt.PolicyName
	}
	title := pt.PolicyTitle
	if title == "" {
		title = pt.PolicyName
	}
	metadata := Metadata{
		Category:    category,
		Description: desc,
		Severity:    severity,
		Title:       pt.PolicyName,
		Id:          pt.GeneratePolicyID(),
	}
	metadata.Severity = "sd"

	data, err := yaml.Marshal(&metadata)

	if err != nil {
		log.Fatal(err)
	}

	metadataPath := strings.Split(pt.PolicyDir, pt.CheckType)[0] + "/matadata.yaml"
	err2 := os.WriteFile(metadataPath, data, 777)

	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println("data written")
	return nil
}

func (pt *PolicyTemplate) CreateDirectoryStructure() error {
	// full directory path
	pt.PolicyDir += "/policies/" + pt.PolicyType + "/" + pt.PolicyName + "/" + pt.CheckType + "/tests"
	if err := os.MkdirAll(pt.PolicyDir, os.ModePerm); err != nil {
		return err
	} else {
		fmt.Println("created: ", pt.PolicyDir)
	}

	return nil
}

func (pt *PolicyTemplate) GeneratePolicyID() string {
	//example:  opal-1238741-s3-block-public-access
	return pt.PolicyType + "-someidval123-" + pt.PolicyName
}
