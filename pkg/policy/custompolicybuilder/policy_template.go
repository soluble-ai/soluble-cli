package custompolicybuilder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"gopkg.in/yaml.v3"
)

type PolicyTemplate struct {
	Name      string
	CheckType string
	Dir       string
	Tool      string
	Desc      string
	Title     string
	Severity  string
	Category  string
	RsrcType  string
	Provider  string
	InputPath string
}

var categories = []string{
	"iam",
	"storage",
	"network",
	"loadbalancers",
	"compute",
	"certs",
	"secrets",
	"encryption",
	"tls",
	"logging",
	"dns",
	"queues",
	"containers",
	"monitoring",
	"tools",
	"security",
	"general",
	"backup & recovery",
}

var severity = []string{
	"info",
	"low",
	"medium",
	"high",
	"critical",
}

var providers = []string{
	"aws",
	"gcp",
	"azure",
	"kubernetes",
	"github",
	"oracle",
}

func getCheckTypes() []string {
	checkTypes := policy.InputTypeForTarget
	keys := make([]string, 0, len(checkTypes))
	for key := range checkTypes {
		keys = append(keys, string(key))
	}
	return keys
}

func (pt *PolicyTemplate) PromptInput() error {
	var qs = []*survey.Question{
		{
			Name: "InputPath",
			Prompt: &survey.Input{
				Message: "Policies directory path.",
				Default: "policies"},
			Validate: pt.validatePolicyDirectory(),
		},
		{
			Name: "provider",
			Prompt: &survey.Select{
				Message: "Select provider:",
				Options: providers,
			},
		},
		{
			Name: "checkType",
			Prompt: &survey.Select{
				Message: "Select target:",
				Options: getCheckTypes(),
				Help:    "type of check (think: what is being examined?)",
			},
		},
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "policy name",
				Help:    "Policy Name may consist of lowercase letters, numbers and underscores. EG: my_policy_1"},
			Validate: pt.validatePolicyName(),
		},
		{
			Name: "title",
			Prompt: &survey.Input{
				Message: "Title",
				Help:    "Max length is 57",
			},
			Validate: survey.ComposeValidators(survey.MinLength(1), survey.MaxLength(57)),
		},
		{
			Name:   "desc",
			Prompt: &survey.Input{Message: "Description"},
		},
		{
			Name: "category",
			Prompt: &survey.Select{
				Message: "Category",
				Options: categories,
				Help:    "functional grouping of the check",
			},
		},
		{
			Name: "rsrcType",
			Prompt: &survey.Input{
				Message: "ResourceType",
				Help:    "For example: aws_s3_bucket",
			},
		},
		{
			Name: "severity",
			Prompt: &survey.Select{
				Message: "Select severity:",
				Options: severity,
			},
		},
	}

	if err := survey.Ask(qs, pt); err != nil {
		return err
	}
	return nil
}

func (pt *PolicyTemplate) validatePolicyName() func(interface{}) error {
	return func(inputName interface{}) error {
		if isValid := regexp.MustCompile(`(^[a-z][a-z0-9_]*$)`).MatchString(inputName.(string)); !isValid {
			return fmt.Errorf("\nname must: \n-start with lowercase letter \n-only contain lowercase letters, numbers and underscored")
		}

		// avoid overwriting existing policy
		path := filepath.Join(pt.Dir, pt.Tool, inputName.(string), pt.CheckType)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return fmt.Errorf("custom policy '%v' with check type '%v' already exists", inputName, pt.CheckType)
		}
		return nil
	}
}

func (pt *PolicyTemplate) createPoliciesDirectoryPrompt(dir, message string) error {
	create := false
	err := survey.AskOne(&survey.Confirm{
		Message: message,
	},
		&create)

	if err != nil {
		return err
	}
	if create {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			// if directory already exists, prompt user to input this path
			// to confirm this is the intended target directory
			return fmt.Errorf("\033[34m %s \033[0m already exists. Input this path to confirm this is the target directory", dir)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		} else {
			log.Infof("created: %s", dir)
		}
		pt.Dir = dir
	} else {
		return fmt.Errorf("provide path to a 'policies' directory")
	}
	return nil
}

func (pt *PolicyTemplate) validatePolicyDirectory() func(interface{}) error {
	return func(inputDir interface{}) error {
		dir := inputDir.(string)
		pt.Dir = dir

		switch {
		case isPoliciesPath(dir):
			// path points to a policies dir
			// check dir exists
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				// if dir doesn't exist offer to create it
				err := pt.createPoliciesDirectoryPrompt(dir,
					dir+" is not an existing policies directory. Create this directory now?")
				if err != nil {
					return err
				}
			}
		case dir == "." || dir == "./":
			// check current dir is named policies
			workingDir, _ := os.Getwd()
			if !isPoliciesPath(workingDir) {
				// offer to create `policies` dir in current dir
				err := pt.createPoliciesDirectoryPrompt("./policies",
					"current directory is not named policies. Create policies directory in current directory?")
				if err != nil {
					return err
				}
			}
		default:
			// path does not point to a policies dir
			// offer to create `policies dir in provided path
			err := pt.createPoliciesDirectoryPrompt(filepath.Join(dir, "policies"),
				dir+" path does not point to a policies directory. Create policies directory here?")
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func isPoliciesPath(path string) bool {
	if path == "policies" {
		return true
	}
	split := strings.LastIndex(path, "/")
	if split == -1 {
		return false
	}
	last := path[split:]
	return last == "/policies"
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
		return err
	}

	err2 := os.WriteFile(metadataPath, data, os.ModePerm)

	if err2 != nil {
		return err2
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
		return err
	}
	return nil
}
