package policy

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"

	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"gopkg.in/yaml.v3"
)

type Metadata map[string]interface{}

func (m Metadata) GetString(key string) string {
	val, _ := m[key].(string)
	return val
}

type Rule struct {
	ID         string
	Path       string
	Metadata   Metadata
	Targets    []Target
	TargetData map[Target]interface{}
}

type Target string

const (
	Terraform      = Target("terraform")
	TerraformPlan  = Target("terraform-plan")
	Cloudformation = Target("cloudformation")
	Kubernetes     = Target("kubernetes")
	Helm           = Target("helm")
	Docker         = Target("docker")
	Secrets        = Target("secrets")
	None           = Target("")
)

var allTargets = []Target{
	Terraform, TerraformPlan, Cloudformation, Kubernetes, Helm, Docker, Secrets,
}

type RuleType interface {
	GetName() string
	GetCode() string
	PrepareRules(rules []*Rule, dest string) error
}

var allRuleTypes = map[string]RuleType{}

type Store struct {
	Dir   string
	Rules map[RuleType][]*Rule
}

func RegisterRuleType(ruleType RuleType) {
	allRuleTypes[ruleType.GetName()] = ruleType
}

func GetRuleType(ruleTypeName string) RuleType {
	return allRuleTypes[ruleTypeName]
}

func (t Target) Path(rule *Rule) string {
	return filepath.Join(rule.Path, string(t))
}

func (m *Store) LoadRules() error {
	m.Rules = make(map[RuleType][]*Rule)
	for _, ruleType := range GetRuleTypes() {
		if err := m.LoadRulesOfType(ruleType); err != nil {
			return err
		}
	}
	return nil
}

func (m *Store) LoadRulesOfType(ruleType RuleType) error {
	ruleTypeDir := filepath.Join(m.Dir, "policies", ruleType.GetName())
	dirs, err := os.ReadDir(ruleTypeDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, ruleDir := range dirs {
		if !ruleDir.IsDir() {
			continue
		}
		dirName := ruleDir.Name()
		if dirName[0] == '.' {
			continue
		}
		_, err := m.LoadSingleRule(ruleType, filepath.Join(ruleTypeDir, dirName))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Store) LoadSingleRule(ruleType RuleType, path string) (*Rule, error) {
	id := fmt.Sprintf("c-%s-%s", ruleType.GetCode(), strings.ReplaceAll(filepath.Base(path), "_", "-"))
	d, err := os.ReadFile(filepath.Join(path, "metadata.yaml"))
	if err != nil {
		return nil, err
	}
	rule := &Rule{
		ID:         id,
		Path:       path,
		TargetData: make(map[Target]interface{}),
	}
	if err := yaml.Unmarshal(d, &rule.Metadata); err != nil {
		return nil, fmt.Errorf("the metadata for %s is invalid - %w", rule.Path, err)
	}
	if rule.Metadata == nil {
		rule.Metadata = make(Metadata)
	}
	rule.Metadata["ruleId"] = rule.ID
	rule.Metadata["sid"] = rule.ID
	log.Debugf("Loaded %s from %s\n", rule.ID, rule.Path)
	m.Rules[ruleType] = append(m.Rules[ruleType], rule)
	for _, target := range allTargets {
		targetDir := filepath.Join(rule.Path, string(target))
		if util.DirExists(targetDir) {
			rule.Targets = append(rule.Targets, target)
		}
	}
	return rule, nil
}

func (m *Store) PrepareRules(dest string) error {
	var err error
	for ruleType, rules := range m.Rules {
		if perr := ruleType.PrepareRules(rules, dest); perr != nil {
			err = multierror.Append(err, perr)
		}
	}
	return err
}

func (m *Store) RuleCount() (count int) {
	for _, rules := range m.Rules {
		count += len(rules)
	}
	return
}

func (m *Store) CreateTarBall(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	w := tar.NewWriter(gz)
	if err := m.writeRules(w); err != nil {
		return err
	}
	if err := m.writeUploadMetadata(w); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	log.Infof("Created tarball with {info:%d} rules", m.RuleCount())
	return gz.Close()
}

func (m *Store) writeUploadMetadata(w *tar.Writer) error {
	env := xcp.GetCIEnv(m.Dir)
	dat, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	h := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "policies-upload-metadata.json",
		Size:     int64(len(dat)),
		ModTime:  time.Now(),
		Mode:     0644,
	}
	if err := w.WriteHeader(h); err != nil {
		return err
	}
	if _, err := w.Write(dat); err != nil {
		return err
	}
	return nil
}

func (m *Store) writeRules(w *tar.Writer) error {
	for _, ruleType := range GetRuleTypes() {
		rules := m.Rules[ruleType]
		for _, rule := range rules {
			log.Infof("Including {info:%s} from {primary:%s}", rule.ID, rule.Path)
			if err := m.writeRuleFiles(w, rule); err != nil {
				return err
			}
			if err := m.writeRuleMetadata(w, rule); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Store) writeRuleMetadata(w *tar.Writer, rule *Rule) error {
	dat, err := yaml.Marshal(rule.Metadata)
	if err != nil {
		return err
	}
	rpath, err := filepath.Rel(m.Dir, rule.Path)
	if err != nil {
		return err
	}
	h := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     fmt.Sprintf("%s/metadata.yaml", rpath),
		Size:     int64(len(dat)),
		ModTime:  time.Now(),
		Mode:     0644,
	}
	if err := w.WriteHeader(h); err != nil {
		return err
	}
	if _, err := w.Write(dat); err != nil {
		return err
	}
	return nil
}

func (m *Store) writeRuleFiles(w *tar.Writer, rule *Rule) error {
	return filepath.Walk(rule.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "metadata.yaml" {
			return nil
		}
		rpath, err := filepath.Rel(m.Dir, path)
		if err != nil {
			return err
		}
		h := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     rpath,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			Mode:     0644,
		}
		if base := filepath.Base(path); base[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			h.Typeflag = tar.TypeDir
			h.Mode = 0755
		}
		if err := w.WriteHeader(h); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
		return nil
	})
}

func (m *Store) GetPolicyUploadMetadata() (map[string]string, error) {
	dat, err := os.ReadFile(filepath.Join(m.Dir, "policies-upload-metadata.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var md map[string]interface{}
	if err := json.Unmarshal(dat, &md); err != nil {
		return nil, err
	}
	res := map[string]string{}
	for k, v := range md {
		if s, ok := v.(string); ok {
			res[k] = s
		}
	}
	return res, err
}

func GetRuleTypes() (res []RuleType) {
	for _, ruleType := range allRuleTypes {
		res = append(res, ruleType)
	}
	// sort so that validate and test run in the same
	// order each time
	sort.Slice(res, func(i, j int) bool {
		return strings.Compare(res[i].GetName(), res[j].GetName()) > 0
	})
	return
}
