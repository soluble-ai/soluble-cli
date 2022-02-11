package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

func DetectPolicy(dir string) (m *Manager, ruleType RuleType, rule *Rule, target Target, err error) {
	// dir may be a root (i.e. it contains a policies directory)
	// policies/<rule-type>
	// policies/<rule_type>/<rule> (contains metadata.yaml)
	// policies/<rule_type>/<rule>/<target>
	if !filepath.IsAbs(dir) {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return
		}
	}
	if util.DirExists(filepath.Join(dir, "policies")) {
		m = NewManager(dir)
		err = m.LoadAllRules()
	} else {
		var absDir string
		absDir, err = filepath.Abs(dir)
		if err != nil {
			log.Errorf("Could not make {info:%s} absolute - {danger:%s}", dir, err)
			return
		}
		elements := strings.Split(absDir, string(os.PathSeparator))
		for i := len(elements) - 1; i > 0; i-- {
			if elements[i] == "policies" {
				m = NewManager(strings.Join(elements[0:i], string(os.PathSeparator)))
				n := len(elements) - i
				if n > 1 {
					ruleType = getRuleType(elements[i+1])
					if ruleType == nil {
						err = fmt.Errorf("unsupported rule type %s", elements[i+1])
						return
					}
				}
				if n > 2 {
					ruleName := elements[i+2]
					rule, err = m.LoadRule(ruleType, strings.Join(elements[0:i+3], string(os.PathSeparator)))
					if err != nil {
						log.Errorf("could not load rule {info:%s} - {danger:%s}", ruleName, err)
						return
					}
				}
				if n > 3 {
					targetName := elements[i+3]
					for _, t := range allTargets {
						if t == Target(targetName) {
							target = t
							break
						}
					}
					if target == "" {
						err = fmt.Errorf("unsupported target %s", targetName)
					}
				}
				if rule == nil && ruleType != nil {
					err = m.LoadRules(ruleType)
				}
				if rule == nil && ruleType == nil {
					err = m.LoadAllRules()
				}
				break
			}
		}
	}
	if m == nil {
		err = fmt.Errorf("%s does not contain policies", dir)
	}
	return
}

func getRuleType(name string) RuleType {
	for _, ruleType := range allRuleTypes {
		if ruleTypeName(ruleType) == name {
			return ruleType
		}
	}
	return nil
}
