package custompolicybuilder

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type PolicyTemplate struct {
	PolicyType string
	CheckType  string
	PolicyName string
	PolicyDir  string
	// optional
	PolicyDesc     string
	PolicyTitle    string
	PolicySeverity string
	PolicyCategory string
	PolicyRsrcType string
}

var SeverityNames = util.NewStringSetWithValues([]string{
	"info", "low", "medium", "high", "critical",
})

func (pt *PolicyTemplate) ValidateCreateInput() error {
	// TODO add validation for optional input
	if isValid := regexp.MustCompile(`^[a-z0-9-]*$`).MatchString(pt.PolicyName); !isValid {
		return fmt.Errorf("invalid policy-name: %v. policy-name must consist only of [a-z0-9-]", pt.PolicyName)
	}

	pt.PolicyType = strings.ToLower(pt.PolicyType)
	if policy.GetRuleType(pt.PolicyType) == nil {
		return fmt.Errorf("invalid policy-type. policy-type is one of: %v", policy.ListRuleTypes())
	}

	pt.CheckType = strings.ToLower(pt.CheckType)
	if !policy.IsTarget(pt.CheckType) {
		return fmt.Errorf("invalid check-type. check-type is one of: %v", policy.ListTargets())
	}

	pt.PolicySeverity = strings.ToLower(pt.PolicySeverity)
	if !SeverityNames.Contains(pt.PolicySeverity) {
		return fmt.Errorf("invalid severity '%v'. severity is one of %v: ", pt.PolicySeverity, SeverityNames.Values())
	}

	if pt.PolicyDir == "policies" {
		if _, err := os.Stat(pt.PolicyDir); os.IsNotExist(err) {
			return fmt.Errorf("could not find '%v' directory in current directory."+
				"\ncreate 'policies' directory or use -d to target existing policies directory", pt.PolicyDir)
		}
	} else {
		dir := "/policies"
		if pt.PolicyDir[len(pt.PolicyDir)-len(dir):] != dir {
			return fmt.Errorf("invalid directory path: %v", pt.PolicyDir+
				"\nprovide path to existing policies directory")
		} else {
			if _, err := os.Stat(pt.PolicyDir); os.IsNotExist(err) {
				return fmt.Errorf("could not find directory: %v", pt.PolicyDir+
					"\ntarget existing policies directory.")
			}
		}
	}
	return nil
}

func (pt *PolicyTemplate) Register(c *cobra.Command) {
	flags := c.Flags()
	flags.StringVar(&pt.PolicyName, "name", "", "name of policy to create")
	_ = c.MarkFlagRequired("name")
	flags.StringVar(&pt.CheckType, "check-type", "", "policy target")
	_ = c.MarkFlagRequired("check-type")
	flags.StringVar(&pt.PolicyType, "type", "", "policy type")
	_ = c.MarkFlagRequired("type")
	// Optional
	flags.StringVarP(&pt.PolicyDir, "directory", "d", "policies", "path to custom policies directory")
	flags.StringVar(&pt.PolicyDesc, "description", "", "policy description")
	flags.StringVar(&pt.PolicyTitle, "title", "", "policy title")
	flags.StringVar(&pt.PolicySeverity, "severity", "medium", "policy severity")
	flags.StringVar(&pt.PolicyCategory, "category", "", "policy category")
	flags.StringVar(&pt.PolicyRsrcType, "resource-type", "", "policy resource type")
}
