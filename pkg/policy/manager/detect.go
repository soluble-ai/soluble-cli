package manager

import (
	"fmt"
	policies "github.com/soluble-ai/soluble-cli/pkg/policy"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

// TODO update from policys to policies
// Finds and loads policies in a directory.  The directory may be any directory
// in the policies tree.  Only policies underneath the directory will be loaded.
// Policy directories have the following layout:
//
// policies/
// policies/<policy-tyoe>
// policies/<policy-type>/<policy> (must contain metadata.yaml)
// policies/<policy-type>/<policy>/<target>
//
// <target> is optional depending on <policy-type>.
func (m *M) DetectPolicy(dir string) error {
	if dir != "" {
		m.Dir = dir
	}
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
		if err := m.LoadPolicies(); err != nil {
			return err
		}
	} else {
		dir := m.Dir
		m.Dir = ""
		m.Policies = nil
		elements := strings.Split(dir, string(os.PathSeparator))
		for i := len(elements) - 1; i > 0; i-- {
			// work backwards through path elements of dir, looking for
			// "policies" directory
			if elements[i] == "policies" {
				var policyType policies.PolicyType
				m.Dir = strings.Join(elements[0:i], string(os.PathSeparator))
				m.Policies = make(map[policies.PolicyType][]*policies.Policy)
				// look at the path elements past "policies" to see where
				// we are
				n := len(elements) - i
				if n > 1 {
					policyType = policies.GetPolicyType(elements[i+1])
					if policyType == nil {
						return fmt.Errorf("unsupported policy type %s", elements[i+1])
					}
				}
				if n > 2 {
					policyPath := strings.Join(elements[0:i+3], string(os.PathSeparator))
					_, err := m.LoadSinglePolicy(policyType, policyPath)
					if err != nil {
						return err
					}
					break
				}
				if policyType != nil {
					if err := m.LoadPoliciesOfType(policyType); err != nil {
						return err
					}
					break
				}
				if err := m.LoadPolicies(); err != nil {
					return err
				}
				break
			}
		}
		if m.Dir == "" {
			return fmt.Errorf("%s is not a policy directory", dir)
		}
	}
	return nil
}
