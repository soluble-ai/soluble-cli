package terraformsettings

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/soluble-ai/soluble-cli/pkg/log"
)

//go:generate go run ../../../gen/gen_terraform_versions.go

//go:embed terraform_versions.txt
var terraformVersions []byte

type providerConfigFile struct {
	TerraformSettings []*TerraformSettings `hcl:"terraform,block"`
	Remain            hcl.Body             `hcl:",remain"`
}

type TerraformSettings struct {
	RequiredVersion *string  `hcl:"required_version"`
	Remain          hcl.Body `hcl:",remain"`
}

func Read(dir string) *TerraformSettings {
	settings := &TerraformSettings{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && dir != path {
			return filepath.SkipDir
		}
		if strings.HasSuffix(path, ".tf") || strings.HasSuffix(path, ".json") {
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			Parse(path, src, settings)
		}
		return nil
	})
	if err != nil {
		log.Warnf("Could not read terraform settings from {warning:%s} - {danger:%s}", dir, err)
	}
	return settings
}

func Parse(filename string, src []byte, settings *TerraformSettings) {
	config := providerConfigFile{}
	if strings.HasSuffix(filename, ".tf") {
		filename += ".hcl"
	}
	err := hclsimple.Decode(filename, src, nil, &config)
	if err != nil {
		fmt.Println(err)
	}
	for _, s := range config.TerraformSettings {
		settings.merge(s)
	}
}

func (t *TerraformSettings) merge(s *TerraformSettings) {
	if s.RequiredVersion != nil {
		t.RequiredVersion = s.RequiredVersion
	}
}

func (t *TerraformSettings) GetTerraformVersion() string {
	if t.RequiredVersion == nil {
		return ""
	}
	c, err := version.NewConstraint(*t.RequiredVersion)
	if err != nil {
		log.Warnf("Invalid {danger:required_version %s} in terraform", *t.RequiredVersion)
		return ""
	}
	s := bufio.NewScanner(bytes.NewBuffer(terraformVersions))
	for s.Scan() {
		v := version.Must(version.NewVersion(s.Text()))
		if c.Check(v) {
			log.Infof("Using terraform version {primary:%s} for init", v.Original())
			return v.Original()
		}
	}
	return ""
}
