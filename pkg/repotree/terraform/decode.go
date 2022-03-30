package terraform

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/zclconf/go-cty/cty"
)

type terraformFile struct {
	Providers []*provider  `hcl:"provider,block"`
	Resources []*resource  `hcl:"resource,block"`
	Modules   []*module    `hcl:"module,block"`
	Terraform []*terraform `hcl:"terraform,block"`
	Remain    hcl.Body     `hcl:",remain"`
}

type provider struct {
	Name   string   `hcl:",label"`
	Alias  string   `hcl:"alias,optional"`
	Remain hcl.Body `hcl:",remain"`
}

type resource struct {
	Type   string   `hcl:",label"`
	Name   string   `hcl:",label"`
	Remain hcl.Body `hcl:",remain"`
}

type module struct {
	Name    string   `hcl:",label"`
	Source  string   `hcl:"source,attr"`
	Version string   `hcl:"version,optional"`
	Remain  hcl.Body `hcl:",remain"`
}

type terraform struct {
	RequiredVersion   string                 `hcl:"required_version,optional"`
	RequiredProviders *requiredProviderBlock `hcl:"required_providers,block"`
	Backend           *backend               `hcl:"backend,block"`
	Remain            hcl.Body               `hcl:",remain"`
}

type requiredProviderBlock struct {
	Body   hcl.Body `hcl:",body"`
	Remain hcl.Body `hcl:",remain"`
}

type requiredProvider struct {
	Alias   string
	Version string
	Source  string
}

type backend struct {
	Type   string   `hcl:",label"`
	Remain hcl.Body `hcl:",remain"`
}

func decode(filename string, src []byte) *terraformFile {
	var file *hcl.File
	var diags hcl.Diagnostics

	switch suffix := strings.ToLower(filepath.Ext(filename)); suffix {
	case ".tf":
		file, diags = hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	case ".json":
		file, diags = json.Parse(src, filename)
	}
	warnErrors(diags)
	if file == nil {
		return nil
	}
	target := &terraformFile{}
	warnErrors(gohcl.DecodeBody(file.Body, nil, target))
	if target.isEmpty() {
		target = nil
	}
	return target
}

func (tf *terraformFile) isEmpty() bool {
	return len(tf.Modules) == 0 && len(tf.Providers) == 0 &&
		len(tf.Resources) == 0 && len(tf.Terraform) == 0
}

func (rpb *requiredProviderBlock) decode() ([]*requiredProvider, hcl.Diagnostics) {
	attrs, diags := rpb.Body.JustAttributes()
	if diags.HasErrors() {
		return nil, diags
	}
	var rps []*requiredProvider
	for name, attr := range attrs {
		rp := &requiredProvider{
			Alias: name,
		}
		// the legacy format is a single-value version constraint only
		if vv, err := attr.Expr.Value(nil); err == nil && vv.Type().IsPrimitiveType() {
			rp.Version = stringValue(attr.Expr)
		} else {
			kvs, diags := hcl.ExprMap(attr.Expr)
			if diags.HasErrors() {
				return nil, diags
			}
			for _, kv := range kvs {
				switch stringValue(kv.Key) {
				case "version":
					rp.Version = stringValue(kv.Value)
				case "source":
					rp.Source = stringValue(kv.Value)
				}
			}
		}
		rps = append(rps, rp)
	}
	return rps, nil
}

func stringValue(expr hcl.Expression) string {
	v, diags := expr.Value(nil)
	if diags.HasErrors() {
		return ""
	}
	if v.Type() != cty.String {
		return ""
	}
	return v.AsString()
}

func warnErrors(diags hcl.Diagnostics) {
	if diags != nil && diags.HasErrors() {
		log.Warnf("{warning:%s}", diags)
	}
}
