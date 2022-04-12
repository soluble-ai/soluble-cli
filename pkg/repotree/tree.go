package repotree

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/repotree/terraform"
)

type Tree struct {
	FileCount                    int                                   `json:"file_count"`
	TerraformLocalResourceCounts map[string]int                        `json:"terraform_local_resource_counts,omitempty"`
	TerraformTopLevelModules     []string                              `json:"terraform_top_level_modules,omitempty"`
	TerraformBackends            []string                              `json:"terraform_backends,omitempty"`
	TerraformExternalModules     map[string]TerraformExternalModuleUse `json:"terraform_external_modules,omitempty"`
	CDKDirectories               []string                              `json:"-"` // omit for now
	Files                        map[string]*File                      `json:"files,omitempty"`
}

type TerraformExternalModuleUse struct {
	Version    string `json:"version,omitempty"`
	UsageCount int    `json:"usage_count"`
}

type File struct {
	Path      string              `json:"-"`
	Modified  bool                `json:"modified,omitempty"`
	Deleted   bool                `json:"deleted,omitempty"`
	Terraform *terraform.Metadata `json:"terraform,omitempty"`
}

func Do(dir string) (*Tree, error) {
	tree := &Tree{
		Files:                        map[string]*File{},
		TerraformLocalResourceCounts: map[string]int{},
	}
	if err := tree.addLsFiles(dir, func(f *File) {}); err != nil {
		return nil, err
	}
	if err := tree.addLsFiles(dir, func(f *File) {
		f.Modified = true
	}, "-m"); err != nil {
		return nil, err
	}
	if err := tree.addLsFiles(dir, func(f *File) {
		f.Deleted = true
	}, "-d"); err != nil {
		return nil, err
	}
	for _, f := range tree.Files {
		var err error
		f.Terraform, err = terraform.Read(filepath.Join(dir, f.Path))
		if err != nil {
			log.Warnf("could not read {warning:%s} - {danger:%s}", f.Path, err)
		}
	}
	tree.summarize()
	return tree, nil
}

func (tree *Tree) addLsFiles(root string, fn func(f *File), args ...string) error {
	c := exec.Command("git", "ls-files", "-z")
	c.Args = append(c.Args, args...)
	c.Dir = root
	c.Stderr = os.Stderr
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return err
	}
	sc := bufio.NewScanner(stdout)
	sc.Split(scanNull)
	for sc.Scan() {
		f := &File{
			Path: sc.Text(),
		}
		tree.Files[f.Path] = f
		fn(f)
	}
	return sc.Err()
}

func scanNull(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func (tree *Tree) summarize() {
	tree.FileCount = len(tree.Files)
	tree.summarizeTerraform()
	tree.summarizeCDK()
}

func (tree *Tree) summarizeCDK() {
	for _, f := range tree.Files {
		if f.Deleted {
			continue
		}
		if filepath.Base(f.Path) == "cdk.json" {
			tree.CDKDirectories = append(tree.CDKDirectories, filepath.Dir(f.Path))
		}
	}
	sort.Strings(tree.CDKDirectories)
}

func (tree *Tree) summarizeTerraform() {
	usedModules := map[string]bool{}
	moduleMetadata := map[string][]*terraform.Metadata{}
	backends := map[string]bool{}
	externalModules := map[terraform.ModuleUse]int{}
	for _, f := range tree.Files {
		if f.Deleted {
			continue
		}
		if f.Terraform != nil {
			dir := filepath.Dir(f.Path)
			for _, mod := range f.Terraform.ModulesUsed {
				if mod.Source != "" {
					if mod.Source[0] == '.' {
						localSource := filepath.Join(dir, mod.Source)
						usedModules[localSource] = true
					} else {
						externalModules[*mod] += 1
					}
				}
			}
			for r, c := range f.Terraform.ResourceCounts {
				tree.TerraformLocalResourceCounts[r] += c
			}
			moduleMetadata[dir] = append(moduleMetadata[dir], f.Terraform)
			if f.Terraform.Settings != nil && f.Terraform.Settings.Backend != "" {
				backends[f.Terraform.Settings.Backend] = true
			}
		}
	}
	for backend := range backends {
		tree.TerraformBackends = append(tree.TerraformBackends, backend)
	}
	sort.Strings(tree.TerraformBackends)
	tree.TerraformExternalModules = make(map[string]TerraformExternalModuleUse)
	for mod, count := range externalModules {
		tree.TerraformExternalModules[mod.Source] = TerraformExternalModuleUse{
			Version:    mod.Version,
			UsageCount: count,
		}
	}
	for mod := range moduleMetadata {
		// It's a top-level terraform module if:
		// 1 - it's not included by any other module in this repository
		// 2 - it either has resources
		// 3 - ... or it uses other modules
		if !usedModules[mod] {
			var hasStuff bool
			for _, t := range moduleMetadata[mod] {
				if len(t.ModulesUsed) > 0 {
					hasStuff = true
					break
				}
				if len(t.ResourceCounts) > 0 {
					hasStuff = true
				}
			}
			if hasStuff {
				tree.TerraformTopLevelModules = append(tree.TerraformTopLevelModules, mod)
			}
		}
		sort.Strings(tree.TerraformTopLevelModules)
	}
}

func (tree *Tree) GetFile(path string) *File {
	return tree.Files[path]
}
