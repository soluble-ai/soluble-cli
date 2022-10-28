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

type Policy struct {
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
	ARM            = Target("arm")
	None           = Target("")
)

var InputTypeForTarget = map[Target]string{
	Terraform:      "tf",
	Cloudformation: "cfn",
	Kubernetes:     "k8s",
	ARM:            "arm",
}

var allTargets = []Target{
	Terraform, TerraformPlan, Cloudformation, Kubernetes, Helm, Docker, Secrets,
}

type PolicyType interface {
	GetName() string
	GetCode() string
	PreparePolicies(policies []*Policy, dest string) error
}

var allPolicyTypes = map[string]PolicyType{}

type Store struct {
	Dir       string
	Policies  map[PolicyType][]*Policy
	PolicyIds map[string]string
}

func RegisterPolicyType(policyType PolicyType) {
	allPolicyTypes[policyType.GetName()] = policyType
}

func GetPolicyType(policyTypeName string) PolicyType {
	return allPolicyTypes[policyTypeName]
}

func (t Target) Path(policy *Policy) string {
	return filepath.Join(policy.Path, string(t))
}

func (m *Store) LoadPolicies() error {
	m.Policies = make(map[PolicyType][]*Policy)
	m.PolicyIds = make(map[string]string)
	for _, policyType := range GetPolicyTypes() {
		if err := m.LoadPoliciesOfType(policyType); err != nil {
			return err
		}
	}
	return nil
}

func (m *Store) LoadPoliciesOfType(policyType PolicyType) error {
	policyTypeDir := filepath.Join(m.Dir, "policies", policyType.GetName())
	dirs, err := os.ReadDir(policyTypeDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, policyDir := range dirs {
		if !policyDir.IsDir() {
			continue
		}
		dirName := policyDir.Name()
		if dirName[0] == '.' {
			continue
		}
		_, err := m.LoadSinglePolicy(policyType, filepath.Join(policyTypeDir, dirName))
		if err != nil {
			return err
		}
	}
	return nil
}

// The resolvePolicyId function resolves the policyId for a PolicyType and path (policy folder name).
// The returned id is a lower case string of the form c-{PolicyTypeCode}-path with all underscores replaced by hyphens
// e.g. for an Opal PolicyType with folder path my_policy the id returned is c-opl-my-policy
// An error is returned if a policy with this id already exists in the stored policies
func (m *Store) resolvePolicyID(policyType PolicyType, path string) (string, error) {
	id := strings.ToLower(fmt.Sprintf("c-%s-%s", policyType.GetCode(), strings.ReplaceAll(filepath.Base(path), "_", "-")))
	if _, exists := m.PolicyIds[id]; exists {
		return "", fmt.Errorf("a policy with id: %s already exists", id)
	}
	m.PolicyIds[id] = id
	return id, nil
}

func (m *Store) LoadSinglePolicy(policyType PolicyType, path string) (*Policy, error) {
	id, err := m.resolvePolicyID(policyType, path)
	if err != nil {
		return nil, err
	}
	d, err := os.ReadFile(filepath.Join(path, "metadata.yaml"))
	if err != nil {
		return nil, err
	}
	policy := &Policy{
		ID:         id,
		Path:       path,
		TargetData: make(map[Target]interface{}),
	}
	if err := yaml.Unmarshal(d, &policy.Metadata); err != nil {
		return nil, fmt.Errorf("the metadata for %s is invalid - %w", policy.Path, err)
	}
	if policy.Metadata == nil {
		policy.Metadata = make(Metadata)
	}
	policy.Metadata["policyId"] = policy.ID
	policy.Metadata["sid"] = policy.ID
	log.Debugf("Loaded %s from %s\n", policy.ID, policy.Path)
	m.Policies[policyType] = append(m.Policies[policyType], policy)
	for _, target := range allTargets {
		targetDir := filepath.Join(policy.Path, string(target))
		if util.DirExists(targetDir) {
			policy.Targets = append(policy.Targets, target)
		}
	}
	return policy, nil
}

func (m *Store) PreparePolicies(dest string) error {
	var err error
	for policyType, policies := range m.Policies {
		if perr := policyType.PreparePolicies(policies, dest); perr != nil {
			err = multierror.Append(err, perr)
		}
	}
	return err
}

func (m *Store) PolicyCount() (count int) {
	for _, policies := range m.Policies {
		count += len(policies)
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
	if err := m.writePolicies(w); err != nil {
		return err
	}
	if err := m.writeUploadMetadata(w); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	log.Infof("Created tarball with {info:%d} policies", m.PolicyCount())
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

func (m *Store) writePolicies(w *tar.Writer) error {
	for _, policyType := range GetPolicyTypes() {
		policies := m.Policies[policyType]
		for _, policy := range policies {
			log.Infof("Including {info:%s} from {primary:%s}", policy.ID, policy.Path)
			if err := m.writePolicyFiles(w, policy); err != nil {
				return err
			}
			if err := m.writePolicyMetadata(w, policy); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Store) writePolicyMetadata(w *tar.Writer, policy *Policy) error {
	dat, err := yaml.Marshal(policy.Metadata)
	if err != nil {
		return err
	}
	policyPath, err := filepath.Rel(m.Dir, policy.Path)
	if err != nil {
		return err
	}
	h := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     fmt.Sprintf("%s/metadata.yaml", policyPath),
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

func (m *Store) writePolicyFiles(w *tar.Writer, policy *Policy) error {
	return filepath.Walk(policy.Path, func(path string, info fs.FileInfo, err error) error {
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

func GetPolicyTypes() (res []PolicyType) {
	for _, policyType := range allPolicyTypes {
		res = append(res, policyType)
	}
	// sort so that validate and test run in the same
	// order each time
	sort.Slice(res, func(i, j int) bool {
		return strings.Compare(res[i].GetName(), res[j].GetName()) > 0
	})
	return
}
