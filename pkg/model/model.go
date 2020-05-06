// Copyright 2020 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/version"
	"github.com/spf13/cobra"
)

type modelLoader struct {
	models      []*Model
	diagnostics hcl.Diagnostics
	parser      *hclparse.Parser
}

type Model struct {
	Command       CommandModel `hcl:"command,block"`
	APIPrefix     string       `hcl:"api_prefix"`
	MinCLIVersion *string      `hcl:"min_cli_version"`
	FileName      string
	Version       string
}

type CommandModel struct {
	Type              string            `hcl:"type,label"`
	Name              string            `hcl:"name,label"`
	Use               *string           `hcl:"use"`
	Short             string            `hcl:"short"`
	Aliases           *[]string         `hcl:"aliases"`
	Path              *string           `hcl:"path"`
	Method            *string           `hcl:"method"`
	Parameters        []*ParameterModel `hcl:"parameter,block"`
	ClusterIDOptional *bool             `hcl:"cluster_id_optional"`
	AuthNotRequired   *bool             `hcl:"auth_not_required"`
	Unauthenticated   *bool             `hcl:"unauthenticated"`
	Result            *ResultModel      `hcl:"result,block"`
	Commands          []*CommandModel   `hcl:"command,block"`
	model             *Model
}

const (
	GetMethod    = "GET"
	PostMethod   = "POST"
	PatchMethod  = "PATCH"
	DeleteMethod = "DELETE"
)

var validMethods = []string{
	GetMethod, PostMethod, PatchMethod, DeleteMethod,
}

type ResultModel struct {
	Path        *[]string          `hcl:"path"`
	Columns     *[]string          `hcl:"columns"`
	WideColumns *[]string          `hcl:"wide_columns"`
	Sort        *[]string          `hcl:"sort_by"`
	Formatters  *map[string]string `hcl:"formatters"`
	LocalAction *string            `hcl:"local_action"`
}

type ParameterModel struct {
	Name         string  `hcl:"name,label"`
	Shorthand    *string `hcl:"shorthand"`
	Usage        *string `hcl:"usage"`
	Required     *bool   `hcl:"required"`
	ContextValue *string `hcl:"context_value"`
	DefaultValue *string `hcl:"default_value"`
	Disposition  *string `hcl:"disposition"`
}

var Models []*Model

func Load(source Source) error {
	log.Debugf("Loading models from {info:%s}", source)
	m := &modelLoader{
		parser: hclparse.NewParser(),
	}
	if err := m.loadModels(source, ""); err != nil {
		return err
	}
	wr := hcl.NewDiagnosticTextWriter(
		os.Stdout,        // writer to send messages to
		m.parser.Files(), // the parser's file cache, for source snippets
		78,               // wrapping width
		true,             // generate colored/highlighted output
	)
	_ = wr.WriteDiagnostics(m.diagnostics)
	if m.diagnostics.HasErrors() {
		return fmt.Errorf("some models have errors")
	}
	for _, model := range m.models {
		if model.MinCLIVersion != nil && !version.IsCompatible(*model.MinCLIVersion) {
			log.Warnf("The model in %s is not compatible with this version of the CLI (require %s)",
				model.FileName, *model.MinCLIVersion)
		}
		if err := model.validate(); err != nil {
			return err
		}
	}
	Models = append(Models, m.models...)
	return nil
}

func (m *modelLoader) loadModels(source Source, dirName string) error {
	dir, err := source.GetFileSystem().Open(dirName)
	if err != nil {
		return err
	}
	defer dir.Close()
	fileInfos, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fileInfo := range fileInfos {
		path := dirName + "/" + fileInfo.Name()
		if fileInfo.IsDir() {
			err := m.loadModels(source, path)
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(fileInfo.Name(), ".hcl") {
			err := m.loadModel(source, path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *modelLoader) loadModel(source Source, name string) error {
	f, err := source.GetFileSystem().Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	src, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	file, diag := m.parser.ParseHCL(src, source.GetPath(name))
	m.diagnostics = m.diagnostics.Extend(diag)
	if !diag.HasErrors() {
		model := &Model{
			FileName: source.GetPath(name),
			Version:  source.GetVersion(name, src),
		}
		diag = gohcl.DecodeBody(file.Body, nil, model)
		m.diagnostics = m.diagnostics.Extend(diag)
		if !diag.HasErrors() {
			m.models = append(m.models, model)
			log.Debugf("%s defines command %s", model.FileName, model.Command.Name)
		}
	}
	return nil
}

func (m *Model) validate() error {
	return m.Command.validate(m)
}

func (cm *CommandModel) GetCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   cm.Name,
		Short: cm.Short,
	}
	if cm.Use != nil {
		c.Use = *cm.Use
	}
	if cm.Aliases != nil {
		c.Aliases = *cm.Aliases
	}
	if cm.GetCommandType().IsGroup() {
		for i := range cm.Commands {
			c.AddCommand(cm.Commands[i].GetCommand())
		}
	} else {
		opts := cm.createOptions()
		c.RunE = func(cmd *cobra.Command, args []string) error {
			return cm.run(opts, cmd, args)
		}
		opts.Register(c)
		cm.createFlags(c)
	}
	return c
}

func (cm *CommandModel) createFlags(c *cobra.Command) {
	for _, p := range cm.Parameters {
		if name := p.getFlagName(); name != "" {
			defaultValue := ""
			if p.DefaultValue != nil {
				defaultValue = *p.DefaultValue
			}
			if p.Shorthand != nil {
				c.Flags().StringP(name, *p.Shorthand, defaultValue, *p.Usage)
			} else {
				c.Flags().String(name, defaultValue, *p.Usage)
			}
			if p.Required != nil && *p.Required {
				_ = c.MarkFlagRequired(name)
			}
		}
	}
}

func (cm *CommandModel) createOptions() options.Interface {
	opts := cm.GetCommandType().createOptions()
	if has, ok := opts.(options.HasClientOpts); ok {
		clientOpts := has.GetClientOpts()
		clientOpts.AuthNotRequired = (cm.Unauthenticated != nil && *cm.Unauthenticated) ||
			(cm.AuthNotRequired != nil && *cm.AuthNotRequired)
		clientOpts.APIPrefix = cm.model.APIPrefix
	}
	if has, ok := opts.(options.HasClusterOpts); ok {
		has.GetClusterOpts().ClusterIDOptional = cm.ClusterIDOptional != nil && *cm.ClusterIDOptional
	}
	if has, ok := opts.(options.HasPrintOpts); ok {
		printOpts := has.GetPrintOpts()
		if cm.Result != nil {
			cm.Result.setPrintOpts(printOpts)
		}
	}
	return opts
}

func (cm *CommandModel) validate(m *Model) error {
	cm.model = m
	if cm.Use != nil && strings.HasPrefix(*cm.Use, cm.Name+" ") {
		return fmt.Errorf("use must begin with with name %s", cm.Name)
	}
	if err := cm.GetCommandType().validate(); err != nil {
		return err
	}
	if !cm.GetCommandType().IsGroup() {
		if cm.Method == nil || !util.StringSliceContains(validMethods, *cm.Method) {
			return fmt.Errorf("invalid method %v must be one of %s", cm.Method, strings.Join(validMethods, " "))
		}
		if cm.Path == nil {
			return fmt.Errorf("path is required here")
		}
	}
	for _, p := range cm.Parameters {
		if err := p.validate(); err != nil {
			return fmt.Errorf("invalid parameter %s: %w", p.Name, err)
		}
	}
	if cm.Result != nil {
		if err := (*cm.Result).validate(); err != nil {
			return err
		}
	}
	for _, scm := range cm.Commands {
		if err := scm.validate(m); err != nil {
			return err
		}
	}
	return nil
}

func (cm *CommandModel) GetCommandType() CommandType {
	return CommandType(cm.Type)
}

func (cm *CommandModel) run(opts options.Interface, cmd *cobra.Command, args []string) error {
	var result *jnode.Node
	contextValues := &ContextValues{}
	if hasClusterOpts, ok := opts.(options.HasClusterOpts); ok {
		clusterID := hasClusterOpts.GetClusterOpts().GetClusterID()
		contextValues.Set("clusterID", clusterID)
		log.Debugf("clusterID = %s", clusterID)
	}
	if hasClientOpts, ok := opts.(options.HasClientOpts); ok {
		clientOpts := hasClientOpts.GetClientOpts()
		contextValues.Set("organizationID", clientOpts.GetOrganization())
		parameters, err := cm.getParameterValues(cmd, contextValues)
		if err != nil {
			return err
		}
		path := cm.getPath(contextValues)
		var apiClient client.Interface
		if cm.Unauthenticated != nil && *cm.Unauthenticated {
			apiClient = clientOpts.GetUnauthenticatedAPIClient()
		} else {
			apiClient = clientOpts.GetAPIClient()
		}
		switch *cm.Method {
		case GetMethod:
			result, err = apiClient.GetWithParams(path, parameters)
		case DeleteMethod:
			result, err = apiClient.Delete(path)
		case PostMethod:
			result, err = apiClient.Post(path, toBody(parameters))
		case PatchMethod:
			result, err = apiClient.Patch(path, toBody(parameters))
		default:
			return fmt.Errorf("unknown method %s", *cm.Method)
		}
		if err != nil {
			return err
		}
	}
	if cm.Result != nil && cm.Result.LocalAction != nil {
		var err error
		result, err = LocalAction(*cm.Result.LocalAction).Run(opts, result)
		if err != nil {
			return err
		}
	}
	if hasPrintOpts, ok := opts.(options.HasPrintOpts); ok {
		hasPrintOpts.GetPrintOpts().PrintResult(result)
	}
	log.Debugf("Command %s successful", cm.Name)
	return nil
}

func toBody(parameters map[string]string) *jnode.Node {
	body := jnode.NewObjectNode()
	for k, v := range parameters {
		body.Put(k, v)
	}
	return body
}

func (cm *CommandModel) getPath(contextValues *ContextValues) string {
	path := *cm.Path
	for k, v := range contextValues.values {
		path = strings.ReplaceAll(path, "{"+k+"}", v)
	}
	return path
}

func (cm *CommandModel) getParameterValues(cmd *cobra.Command, contextValues *ContextValues) (map[string]string, error) {
	values := map[string]string{}
	for _, p := range cm.Parameters {
		var value string
		flagName := p.getFlagName()
		if flagName != "" {
			flag := cmd.Flag(flagName)
			value = flag.Value.String()
		} else {
			v, err := contextValues.Get(*p.ContextValue)
			if err != nil {
				return nil, err
			}
			value = v
		}
		switch p.getDisposition() {
		case ContextDisposition:
			contextValues.Set(p.Name, value)
		default:
			values[p.Name] = value
		}
	}
	return values, nil
}

func (p *ParameterModel) getFlagName() string {
	if p.ContextValue != nil {
		return ""
	}
	w := &bytes.Buffer{}
	c := p.Name
	var wasUpper int
	for i, ch := range c {
		upper := unicode.IsUpper(ch)
		if i > 0 && wasUpper == 0 && upper {
			w.WriteRune('-')
		}
		if ch == '_' {
			w.WriteRune('-')
			wasUpper = -1
		} else {
			w.WriteRune(unicode.ToLower(ch))
			wasUpper = 0
			if upper {
				wasUpper = 1
			}
		}
	}
	return w.String()
}

func (p *ParameterModel) validate() error {
	if p.ContextValue == nil && p.Usage == nil {
		return fmt.Errorf("parameter '%s' must have usage", p.Name)
	}
	if err := p.getDisposition().validate(); err != nil {
		return err
	}
	return nil
}

func (p *ParameterModel) getDisposition() ParameterDisposition {
	if p.Disposition != nil {
		return ParameterDisposition(*p.Disposition)
	}
	return ParameterDisposition("")
}

func (r *ResultModel) validate() error {
	if r.Path != nil && r.Columns == nil {
		return fmt.Errorf("if path is given columns must also be specified")
	}
	if r.Formatters != nil {
		for column := range *r.Formatters {
			if !r.getFormatter(column).isValid() {
				return fmt.Errorf("invalid formatter %s for column %s", r.getFormatter(column), column)
			}
		}
	}
	if r.LocalAction != nil {
		err := LocalAction(*r.LocalAction).validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ResultModel) setPrintOpts(opts *options.PrintOpts) {
	if r.Path != nil {
		opts.Path = *r.Path
		opts.Columns = *r.Columns
	}
	if r.WideColumns != nil {
		opts.WideColumns = *r.WideColumns
	}
	if r.Sort != nil {
		opts.SortBy = *r.Sort
	}
	if r.Formatters != nil {
		for column := range *r.Formatters {
			if opts.Formatters == nil {
				opts.Formatters = make(map[string]options.Formatter)
			}
			opts.Formatters[column] = r.getFormatter(column).getFormatter(opts)
		}
	}
}

func (r *ResultModel) getFormatter(column string) ColumnFormatter {
	if r.Formatters != nil {
		return ColumnFormatter((*r.Formatters)[column])
	}
	return ColumnFormatter("")
}
