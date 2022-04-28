package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

// Finds and loads rules in a directory.  The directory may be any directory
// in the policies tree.  Only rules underneath the directory will be loaded.
// Policy directories have the following layout:
//
// policies/
// policies/<rule-tyoe>
// policies/<rule-type>/<rule> (must contain metadata.yaml)
// policies/<rule-type>/<rule>/<target>
//
// <target> is optional depending on <rule-type>.
func DetectPolicy(dir string) (*Manager, error) {
	if !filepath.IsAbs(dir) {
		var err error
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
	}
	var m *Manager
	if util.DirExists(filepath.Join(dir, "policies")) {
		m = NewManager(dir)
		if err := m.LoadAllRules(); err != nil {
			return nil, err
		}
	} else {
		elements := strings.Split(dir, string(os.PathSeparator))
		for i := len(elements) - 1; i > 0; i-- {
			// work backwards through path elements of dir, looking for
			// "policies" directory
			if elements[i] == "policies" {
				var ruleType RuleType
				m = NewManager(strings.Join(elements[0:i], string(os.PathSeparator)))
				// look at the path elements past "policies" to see where
				// we are
				n := len(elements) - i
				if n > 1 {
					ruleType = allRuleTypes[elements[i+1]]
					if ruleType == nil {
						return nil, fmt.Errorf("unsupported rule type %s", elements[i+1])
					}
				}
				if n > 2 {
					rulePath := strings.Join(elements[0:i+3], string(os.PathSeparator))
					_, err := m.LoadRule(ruleType, rulePath)
					if err != nil {
						return nil, err
					}
					break
				}
				if ruleType != nil {
					if err := m.LoadRules(ruleType); err != nil {
						return nil, err
					}
					break
				}
				if err := m.LoadAllRules(); err != nil {
					return nil, err
				}
				break
			}
		}
	}
	if m == nil {
		return nil, fmt.Errorf("%s is not a policy directory", dir)
	}
	return m, m.ValidateRules()
}
