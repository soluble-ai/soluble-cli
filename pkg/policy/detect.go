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
func (m *Manager) DetectPolicy() error {
	if !filepath.IsAbs(m.Dir) {
		dir, err := filepath.Abs(m.Dir)
		if err != nil {
			return err
		}
		m.Dir = dir
	}
	if !util.DirExists(m.Dir) {
		return fmt.Errorf("%s is not a directory", m.Dir)
	}
	if util.DirExists(filepath.Join(m.Dir, "policies")) {
		if err := m.LoadRules(); err != nil {
			return err
		}
	} else {
		dir := m.Dir
		m.Dir = ""
		m.Rules = nil
		elements := strings.Split(dir, string(os.PathSeparator))
		for i := len(elements) - 1; i > 0; i-- {
			// work backwards through path elements of dir, looking for
			// "policies" directory
			if elements[i] == "policies" {
				var ruleType RuleType
				m.Dir = strings.Join(elements[0:i], string(os.PathSeparator))
				m.Rules = make(map[RuleType][]*Rule)
				// look at the path elements past "policies" to see where
				// we are
				n := len(elements) - i
				if n > 1 {
					ruleType = allRuleTypes[elements[i+1]]
					if ruleType == nil {
						return fmt.Errorf("unsupported rule type %s", elements[i+1])
					}
				}
				if n > 2 {
					rulePath := strings.Join(elements[0:i+3], string(os.PathSeparator))
					_, err := m.loadRule(ruleType, rulePath)
					if err != nil {
						return err
					}
					break
				}
				if ruleType != nil {
					if err := m.loadRules(ruleType); err != nil {
						return err
					}
					break
				}
				if err := m.LoadRules(); err != nil {
					return err
				}
				break
			}
		}
		if m.Dir == "" {
			return fmt.Errorf("%s is not a policy directory", dir)
		}
	}
	_, err := m.ValidateRules()
	return err
}
