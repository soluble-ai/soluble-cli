package custompolicybuilder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/policy"

	"github.com/AlecAivazis/survey/v2"
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
			Name: "provider",
			Prompt: &survey.Select{
				Message: "Select provider:",
				Options: providers,
			},
		},
		{
			Name: "dir",
			Prompt: &survey.Input{
				Message: "Policies directory path",
				Default: "policies"},
			Validate: validatePolicyDirectory(),
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

	if err := survey.Ask(qs, pt); err == nil {
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
func validatePolicyDirectory() func(interface{}) error {
	return func(inputDir interface{}) error {
		dir := inputDir.(string)
		if inputDir == "policies" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("could not find '%v' directory in current directory."+
					"\ncreate 'policies' directory or use -d to target an existing policies directory", dir)
			}
		} else {
			split := strings.LastIndex(dir, "/")
			if split == -1 {
				return fmt.Errorf("invalid directory: %v", dir+
					"\ntarget an existing policies directory.")
			}
			last := dir[split:]
			if last != "/policies" {
				return fmt.Errorf("invalid directory path: %v", dir+
					"\nprovide path to existing policies directory")
			} else {
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					return fmt.Errorf("could not find directory: %v", dir+
						"\ntarget an existing policies directory.")
				}
			}
		}
		return nil
	}
}
