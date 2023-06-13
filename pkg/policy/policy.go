package policy

import (
	"archive/zip"
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

type ConvertedOpalPolicies struct {
	ConvertedPolicies []string `json:"convertedPolicies"`
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
	Policies       = "policies"
)

var InputTypeForTarget = map[Target]string{
	Terraform:      "tf",
	Cloudformation: "cfn",
	Kubernetes:     "k8s",
	ARM:            "arm",
}

var allTargets = []Target{
	Terraform, TerraformPlan, Cloudformation, Kubernetes, Helm, Docker, Secrets, ARM,
}

type PolicyType interface {
	GetName() string
	GetCode() string
	PreparePolicies(policies []*Policy, dest string) error
}

var allPolicyTypes = map[string]PolicyType{}

type Store struct {
	Dir                    string
	Policies               map[PolicyType][]*Policy
	PolicyIds              map[string]string
	SkipPolicyIDResolution bool
	ConvertedPoliciesFile  string
}

func NewStore(dir string, skipPolicyIDResolution bool) *Store {
	return &Store{
		Dir:                    dir,
		Policies:               make(map[PolicyType][]*Policy),
		PolicyIds:              make(map[string]string),
		SkipPolicyIDResolution: skipPolicyIDResolution,
	}
}

func NewDownloadStore(dir string) *Store {
	return NewStore(dir, true)
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
	d, err := os.ReadFile(filepath.Join(path, "metadata.yaml"))
	if err != nil {
		return nil, err
	}
	policy := &Policy{
		Path:       path,
		TargetData: make(map[Target]interface{}),
	}
	if err := yaml.Unmarshal(d, &policy.Metadata); err != nil {
		return nil, fmt.Errorf("the metadata for %s is invalid - %w", policy.Path, err)
	}
	if policy.Metadata == nil {
		policy.Metadata = make(Metadata)
	}
	if m.SkipPolicyIDResolution {
		policy.ID = policy.Metadata.GetString("policyId")
	} else {
		id, err := m.resolvePolicyID(policyType, path)
		if err != nil {
			return nil, err
		}
		policy.ID = id
		policy.Metadata["policyId"] = policy.ID
		policy.Metadata["sid"] = policy.ID
	}
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

func (m *Store) CreateZipArchive(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	zipWriter := zip.NewWriter(f)
	if err := m.writePolicies(zipWriter); err != nil {
		return err
	}
	if err := m.writeUploadMetadata(zipWriter); err != nil {
		return err
	}
	if m.ConvertedPoliciesFile != "" {
		if err := m.writeOpalPolicyTracker(zipWriter); err != nil {
			return err
		}
	}

	zipWriter.Flush()
	if err := zipWriter.Close(); err != nil {
		return err
	}
	log.Infof("Created zip file with {info:%d} policies", m.PolicyCount())
	return nil
}

func (m *Store) writeUploadMetadata(zipWriter *zip.Writer) error {
	env := xcp.GetCIEnv(m.Dir)
	dat, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	header := &zip.FileHeader{
		Name:               "policies-upload-metadata.json",
		Modified:           time.Now(),
		Method:             zip.Deflate,
		UncompressedSize64: uint64(len(dat)),
	}
	header.SetMode(0644)
	ioWriter, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := ioWriter.Write(dat); err != nil {
		return err
	}
	return nil
}

func (m *Store) writeOpalPolicyTracker(zipWriter *zip.Writer) error {

	opalPolicyMappings, err := os.ReadFile(m.ConvertedPoliciesFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	duplicateTrackingFile, err := m.getPublishedPolicyMappings(opalPolicyMappings)
	if err != nil {
		return err
	}
	header := &zip.FileHeader{
		Name:               fmt.Sprintf("opal-duplicates.json"),
		UncompressedSize64: uint64(len(duplicateTrackingFile.ConvertedPolicies)),
		Modified:           time.Now(),
		Method:             zip.Deflate,
	}
	header.SetMode(0644)
	ioWriter, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	dat, _ := json.Marshal(duplicateTrackingFile)
	if _, err := ioWriter.Write(dat); err != nil {
		return err
	}
	return nil
}

func (m *Store) writePolicies(zipWriter *zip.Writer) error {
	_, err := m.writePolicyRootPath(zipWriter)
	if err != nil {
		return err
	}
	for _, policyType := range GetPolicyTypes() {
		policies := m.Policies[policyType]
		if policies != nil {
			_, err = m.writePolicyTypePath(zipWriter, policyType)
			if err != nil {
				return err
			}
		}
		for _, policy := range policies {
			log.Infof("Including {info:%s} from {primary:%s}", policy.ID, policy.Path)
			if err := m.writePolicyFiles(zipWriter, policy); err != nil {
				return err
			}
			if err := m.writePolicyMetadata(zipWriter, policy); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Store) writePolicyMetadata(zipWriter *zip.Writer, policy *Policy) error {
	dat, err := yaml.Marshal(policy.Metadata)
	if err != nil {
		return err
	}
	policyPath, err := filepath.Rel(m.Dir, policy.Path)
	if err != nil {
		return err
	}
	header := &zip.FileHeader{
		Name:               fmt.Sprintf("%s/metadata.yaml", policyPath),
		UncompressedSize64: uint64((len(dat))),
		Modified:           time.Now(),
		Method:             zip.Deflate,
	}
	header.SetMode(0644)
	ioWriter, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := ioWriter.Write(dat); err != nil {
		return err
	}
	return nil
}

func (m *Store) writePolicyFiles(zipWriter *zip.Writer, policy *Policy) error {
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
		header, err := zip.FileInfoHeader(info)
		header.Name = rpath
		header.Method = zip.Deflate
		if err != nil {
			return err
		}
		if info.IsDir() {
			rpath = fmt.Sprintf("%s/", rpath)
			header.Name = rpath
			header.SetMode(0755)
			_, err = zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
		} else if base := filepath.Base(path); base[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		} else {
			policyFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer policyFile.Close()
			header.SetMode(0644)
			ioWriter, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			if _, err := io.Copy(ioWriter, policyFile); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *Store) GetPolicyUploadMetadata(filename string) (map[string]string, error) {
	dat, err := os.ReadFile(filepath.Join(m.Dir, filename))
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

func (m *Store) writePolicyRootPath(zipWriter *zip.Writer) (io.Writer, error) {
	path := filepath.Dir(fmt.Sprintf("%s/%s", m.Dir, Policies))
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	header, _ := zip.FileInfoHeader(fileInfo)
	header.Name = fmt.Sprintf("%s/", Policies)
	header.SetMode(0755)
	header.Method = zip.Deflate
	return zipWriter.CreateHeader(header)
}

func (m *Store) writePolicyTypePath(zipWriter *zip.Writer, policyType PolicyType) (io.Writer, error) {
	path := filepath.Dir(fmt.Sprintf("%s/%s/%s", m.Dir, Policies, policyType.GetName()))
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return nil, err
	}
	header.Name = fmt.Sprintf("%s/%s/", Policies, policyType.GetName())
	header.SetMode(0755)
	header.Method = zip.Deflate
	return zipWriter.CreateHeader(header)
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

func (m *Store) getPublishedPolicyMappings(opalPolicyMappings []byte) (*ConvertedOpalPolicies, error) {
	var convertedPolicies map[string][]string
	if err := json.Unmarshal(opalPolicyMappings, &convertedPolicies); err != nil {
		return nil, err
	}
	duplicateTrackingFile := &ConvertedOpalPolicies{}
	laceworkPolicies := m.Policies[GetPolicyType("opal")]
	for trackedOpalPolicy, oldTrackedPolicy := range convertedPolicies {
		for _, publishedOpalPolicy := range laceworkPolicies {
			if filepath.Base(publishedOpalPolicy.Path) == trackedOpalPolicy {
				duplicateTrackingFile.ConvertedPolicies = append(duplicateTrackingFile.ConvertedPolicies, oldTrackedPolicy...)
			}
		}
	}
	return duplicateTrackingFile, nil
}
