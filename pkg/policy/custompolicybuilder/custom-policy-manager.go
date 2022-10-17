package custompolicybuilder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
)

type PolicyTemplate struct {
	Name      string
	CheckType string
	Dir       string
	Type      string
	Desc      string
	Title     string
	Severity  string
	Category  string
	RsrcType  string
}

func (pt *PolicyTemplate) PromptInput() error {
	var qs = []*survey.Question{
		{
			Name: "dir",
			Prompt: &survey.Input{
				Message: "Policies directory path",
				Default: "policies"},
			Validate: pt.validatePolicyDirectory(),
		},
		{
			Name: "checkType",
			Prompt: &survey.Select{
				Message: "Select target:",
				Options: []string{"terraform", "cloudformation", "kubernetes", "arm"},
			},
		},
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "policy name",
				Help:    "Example policy name: my_policy_1"},
			Validate: pt.validatePolicyName(),
		},
		{
			Name:   "title",
			Prompt: &survey.Input{Message: "Title"},
		},
		{
			Name:   "desc",
			Prompt: &survey.Input{Message: "Description"},
		},
		{
			Name:   "category",
			Prompt: &survey.Input{Message: "Category"},
		},
		{
			Name:   "rsrcType",
			Prompt: &survey.Input{Message: "ResourceType"},
		},
		{
			Name: "severity",
			Prompt: &survey.Select{
				Message: "Select severity:",
				Options: []string{"info", "low", "medium", "high", "critical"},
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
		path := filepath.Join(pt.Dir, pt.Type, inputName.(string), pt.CheckType)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return fmt.Errorf("custom policy '%v' with check type '%v' already exists", inputName, pt.CheckType)
		}
		return nil
	}

}
func (pt *PolicyTemplate) validatePolicyDirectory() func(interface{}) error {
	return func(inputDir interface{}) error {
		dir := inputDir.(string)
		if inputDir == "policies" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("could not find '%v' directory in current directory."+
					"\ncreate 'policies' directory or use -d to target an existing policies directory", dir)
			}
		} else {
			pdir := "/policies"
			if dir[len(dir)-len(pdir):] != pdir {
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
