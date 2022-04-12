package terraform

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type Metadata struct {
	Providers      []*Provider    `json:"providers,omitempty"`
	Settings       *Settings      `json:"settings,omitempty"`
	ModulesUsed    []*ModuleUse   `json:"modules_used,omitempty"`
	ResourceCounts map[string]int `json:"resources"`
}

type Provider struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"`
}

type Settings struct {
	RequiredVersion   string              `json:"required_version,omitempty"`
	RequiredProviders []*RequiredProvider `json:"required_providers,omitempty"`
	Backend           string              `json:"backend,omitempty"`
}

type RequiredProvider struct {
	Alias   string `json:"alias"`
	Version string `json:"version,omitempty"`
	Source  string `json:"source,omitempty"`
}

type ModuleUse struct {
	Source  string `json:"source"`
	Version string `json:"version,omitempty"`
}

func Read(path string) (*Metadata, error) {
	if !strings.HasSuffix(path, ".tf") {
		return nil, nil
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tf := decode(path, src)
	if tf == nil {
		return nil, nil
	}
	var m *Metadata
	if tf != nil {
		m = &Metadata{
			ResourceCounts: map[string]int{},
		}
		for _, r := range tf.Resources {
			m.ResourceCounts[r.Type]++
		}
		for _, p := range tf.Providers {
			m.Providers = append(m.Providers, &Provider{Name: p.Name, Alias: p.Alias})
		}
		modules := map[string]bool{}
		for _, mod := range tf.Modules {
			key := fmt.Sprintf("%s:%s", mod.Source, mod.Version)
			if !modules[key] {
				modules[key] = true
				m.ModulesUsed = append(m.ModulesUsed, &ModuleUse{
					Source:  mod.Source,
					Version: mod.Version,
				})
			}
		}
		sort.Slice(m.ModulesUsed, func(i, j int) bool {
			return strings.Compare(m.ModulesUsed[i].Source, m.ModulesUsed[j].Source) < 0
		})
		if len(tf.Terraform) > 0 {
			m.Settings = &Settings{}
			for _, t := range tf.Terraform {
				if t.RequiredVersion != "" {
					m.Settings.RequiredVersion = t.RequiredVersion
				}
				if t.RequiredProviders != nil {
					rps, err := t.RequiredProviders.decode()
					if err != nil {
						return nil, err
					}
					for _, rp := range rps {
						m.Settings.RequiredProviders = append(m.Settings.RequiredProviders,
							&RequiredProvider{
								Alias:   rp.Alias,
								Source:  rp.Source,
								Version: rp.Version,
							})
					}
				}
				if t.Backend != nil {
					m.Settings.Backend = t.Backend.Type
				}
			}
		}
	}
	return m, nil
}
