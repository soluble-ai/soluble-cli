package custompolicybuilder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type PolicyTemplate struct {
	Type      string
	CheckType string
	Name      string
	Dir       string
	// optional
	Desc     string
	Title    string
	Severity string
	Category string
	RsrcType string
}

var SeverityNames = util.NewStringSetWithValues([]string{
	"info", "low", "medium", "high", "critical",
})

func (pt *PolicyTemplate) ValidateCreateInput() error {
	// TODO add validation for optional input
	if isValid := regexp.MustCompile(`^[a-z0-9_]*$`).MatchString(pt.Name); !isValid {
		return fmt.Errorf("invalid name: %v. name must consist only of [a-z0-9-]", pt.Name)
	}

	pt.Type = strings.ToLower(pt.Type)
	if policy.GetRuleType(pt.Type) == nil {
		return fmt.Errorf("invalid type. type is one of: %v", policy.ListRuleTypes())
	}

	pt.CheckType = strings.ToLower(pt.CheckType)
	if !policy.IsTarget(pt.CheckType) {
		return fmt.Errorf("invalid check-type. check-type is one of: %v", policy.ListTargets())
	}

	pt.Severity = strings.ToLower(pt.Severity)
	if !SeverityNames.Contains(pt.Severity) {
		return fmt.Errorf("invalid severity '%v'. severity is one of: %v", pt.Severity, SeverityNames.Values())
	}

	if pt.Dir == "policies" {
		if _, err := os.Stat(pt.Dir); os.IsNotExist(err) {
			return fmt.Errorf("could not find '%v' directory in current directory."+
				"\ncreate 'policies' directory or use -d to target an existing policies directory", pt.Dir)
		}
	} else {
		dir := "/policies"
		if pt.Dir[len(pt.Dir)-len(dir):] != dir {
			return fmt.Errorf("invalid directory path: %v", pt.Dir+
				"\nprovide path to existing policies directory")
		} else {
			if _, err := os.Stat(pt.Dir); os.IsNotExist(err) {
				return fmt.Errorf("could not find directory: %v", pt.Dir+
					"\ntarget an existing policies directory.")
			}
		}
	}
	path := filepath.Join(pt.Dir, pt.Type, pt.Name, pt.CheckType)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("custom policy '%v' with check type '%v' already exists in directory '%v'",
			pt.Name, pt.CheckType, path)
	}
	return nil
}

func (pt *PolicyTemplate) Register(c *cobra.Command) {
	flags := c.Flags()
	flags.StringVar(&pt.Name, "name", "", "name of policy to create")
	_ = c.MarkFlagRequired("name")
	flags.StringVar(&pt.CheckType, "check-type", "", " target")
	_ = c.MarkFlagRequired("check-type")
	flags.StringVar(&pt.Type, "type", "", "policy type")
	_ = c.MarkFlagRequired("type")
	// Optional
	flags.StringVarP(&pt.Dir, "directory", "d", "policies", "path to custom policies directory")
	flags.StringVar(&pt.Desc, "description", "", "policy description")
	flags.StringVar(&pt.Title, "title", "", "policy title")
	flags.StringVar(&pt.Severity, "severity", "medium", "policy severity")
	flags.StringVar(&pt.Category, "category", "", "policy category")
	flags.StringVar(&pt.RsrcType, "resource-type", "", "policy resource type")
}
