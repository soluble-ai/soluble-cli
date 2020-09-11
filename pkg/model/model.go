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
	"strings"
	"unicode"

	"github.com/hashicorp/hcl/v2"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Model struct {
	Command       CommandModel `hcl:"command,block"`
	APIPrefix     string       `hcl:"api_prefix"`
	MinCLIVersion *string      `hcl:"min_cli_version"`
	FileName      string
	Version       string
	Source        Source
	diagnostics   hcl.Diagnostics
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
	ParameterNames    *[]string         `hcl:"parameters"`
	ClusterIDOptional *bool             `hcl:"cluster_id_optional"`
	AuthNotRequired   *bool             `hcl:"auth_not_required"`
	Unauthenticated   *bool             `hcl:"unauthenticated"`
	DefaultTimeout    *int              `hcl:"default_timeout"`
	Result            *ResultModel      `hcl:"result,block"`
	Commands          []*CommandModel   `hcl:"command,block"`
	ParameterDefs     *ParameterDefs    `hcl:"parameter_defs,block"`
	parameters        []*ParameterModel
	model             *Model
	parent            *CommandModel
}

type ParameterDefs struct {
	Parameters []*ParameterModel `hcl:"parameter,block"`
}

type ResultModel struct {
	Path                    *[]string          `hcl:"path"`
	TruncationIndicatorPath *[]string          `hcl:"truncation_indicator"`
	Columns                 *[]string          `hcl:"columns"`
	WideColumns             *[]string          `hcl:"wide_columns"`
	Sort                    *[]string          `hcl:"sort_by"`
	Formatters              *map[string]string `hcl:"formatters"`
	ComputedColumns         *map[string]string `hcl:"computed_columns"`
	LocalAction             *string            `hcl:"local_action"`
	DiffColumn              *string            `hcl:"diff_column"`
	VersionColumn           *string            `hcl:"version_column"`
	DefaultOutputFormat     *string            `hcl:"default_output_format"`
}

type ParameterModel struct {
	Name         string  `hcl:"name,label"`
	Shorthand    *string `hcl:"shorthand"`
	Usage        *string `hcl:"usage"`
	Required     *bool   `hcl:"required"`
	BooleanFlag  *bool   `hcl:"boolean"`
	LiteralValue *string `hcl:"literal_value"`
	ContextValue *string `hcl:"context_value"`
	DefaultValue *string `hcl:"default_value"`
	Disposition  *string `hcl:"disposition"`
}

var Models []*Model

func (m *Model) validate() error {
	return m.Command.validate(m, nil)
}

func (cm *CommandModel) GetCommand() Command {
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
	command := cm.GetCommandType().makeCommand(c, cm)
	if !cm.GetCommandType().IsGroup() {
		c.RunE = func(cmd *cobra.Command, args []string) error {
			return cm.run(command, cmd, args)
		}
	}
	cm.createFlags(c)

	return command
}

func (cm *CommandModel) createFlags(c *cobra.Command) {
	for _, p := range cm.parameters {
		if name := p.getFlagName(); name != "" {
			defaultValue := ""
			if p.DefaultValue != nil {
				defaultValue = *p.DefaultValue
			}
			switch {
			case p.BooleanFlag != nil && *p.BooleanFlag:
				if p.Shorthand != nil {
					c.Flags().BoolP(name, *p.Shorthand, defaultValue == "true", *p.Usage)
				} else {
					c.Flags().Bool(name, defaultValue == "true", *p.Usage)
				}
			case p.Shorthand != nil:
				c.Flags().StringP(name, *p.Shorthand, defaultValue, *p.Usage)
			default:
				c.Flags().String(name, defaultValue, *p.Usage)
			}
			if p.Required != nil && *p.Required {
				_ = c.MarkFlagRequired(name)
			}
		}
	}
}

func (cm *CommandModel) validate(m *Model, parent *CommandModel) error {
	cm.model = m
	cm.parent = parent
	if cm.Use != nil && strings.HasPrefix(*cm.Use, cm.Name+" ") {
		return fmt.Errorf("command %s: use must begin with with command name", cm.Name)
	}
	if err := cm.GetCommandType().validate(); err != nil {
		return fmt.Errorf("command %s: invalid command type: %w", cm.Name, err)
	}
	if !cm.GetCommandType().IsGroup() {
		if cm.Method == nil || !util.StringSliceContains(validMethods, *cm.Method) {
			return fmt.Errorf("command %s: invalid method %v must be one of %s", cm.Name, cm.Method, strings.Join(validMethods, " "))
		}
		if cm.Path == nil {
			return fmt.Errorf("command %s: path is required", cm.Name)
		}
	}
	for _, p := range cm.Parameters {
		if err := p.validate(); err != nil {
			return fmt.Errorf("command %s: invalid parameter %s: %w", cm.Name, p.Name, err)
		}
	}
	if cm.Result != nil {
		if err := (*cm.Result).validate(); err != nil {
			return fmt.Errorf("invalid result for command %s: %w", cm.Name, err)
		}
	}
	for _, scm := range cm.Commands {
		if err := scm.validate(m, cm); err != nil {
			return fmt.Errorf("command %s: %w", cm.Name, err)
		}
	}
	if cm.ParameterDefs != nil {
		if err := (*cm.ParameterDefs).validate(); err != nil {
			return fmt.Errorf("command %s: invalid parameter definition: %w", cm.Name, err)
		}
	}
	cm.parameters = cm.Parameters
	if cm.ParameterNames != nil {
		for _, name := range *cm.ParameterNames {
			p := cm.getDefinedParameter(name)
			if p == nil {
				return fmt.Errorf("command %s: no defined parameter %s", cm.Name, name)
			}
			cm.parameters = append(cm.parameters, p)
		}
	}
	seen := map[string]bool{}
	for _, p := range cm.parameters {
		if seen[p.Name] {
			return fmt.Errorf("command %s: duplicate parameter %s", cm.Name, p.Name)
		}
		seen[p.Name] = true
	}
	return nil
}

func (cm *CommandModel) getDefinedParameter(name string) *ParameterModel {
	if cm.ParameterDefs != nil {
		for _, p := range cm.ParameterDefs.Parameters {
			if p.Name == name {
				return p
			}
		}
	}
	if cm.parent != nil {
		return cm.parent.getDefinedParameter(name)
	}
	return nil
}

func (cm *CommandModel) GetCommandType() CommandType {
	return CommandType(cm.Type)
}

func (cm *CommandModel) run(command Command, cmd *cobra.Command, args []string) error {
	var result *jnode.Node
	contextValues := NewContextValues()
	command.SetContextValues(contextValues.values)
	parameters, err := cm.processParameters(cmd, contextValues)
	if err != nil {
		return err
	}
	path := cm.getPath(contextValues)
	var apiClient client.Interface
	if cm.Unauthenticated != nil && *cm.Unauthenticated {
		apiClient = command.GetUnauthenticatedAPIClient()
	} else {
		apiClient = command.GetAPIClient()
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
		panic(fmt.Errorf("unknown method %s", *cm.Method))
	}
	if err != nil {
		return err
	}
	if cm.Result != nil && cm.Result.LocalAction != nil {
		var err error
		result, err = LocalActionType(*cm.Result.LocalAction).Run(command, result)
		if err != nil {
			return err
		}
	}
	command.PrintResult(result)
	if cm.Result != nil && cm.Result.TruncationIndicatorPath != nil {
		if err := warnIfTruncated(result, *cm.Result.TruncationIndicatorPath); err != nil {
			return err
		}
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

func (cm *CommandModel) processParameters(cmd *cobra.Command, contextValues *ContextValues) (map[string]string, error) {
	values := map[string]string{}
	for _, p := range cm.parameters {
		var value string
		flagName := p.getFlagName()
		switch {
		case flagName != "":
			flag := cmd.Flag(flagName)
			value = flag.Value.String()
		case p.LiteralValue != nil:
			value = *p.LiteralValue
		default:
			v, err := contextValues.Get(*p.ContextValue)
			if err != nil {
				return nil, err
			}
			value = v
		}
		switch p.getDisposition() {
		case ContextDisposition:
			contextValues.values[p.Name] = value
		default:
			values[p.Name] = value
		}
	}
	return values, nil
}

func (d *ParameterDefs) validate() error {
	for _, p := range d.Parameters {
		if err := p.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (p *ParameterModel) getFlagName() string {
	if p.ContextValue != nil || p.LiteralValue != nil {
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
	if p.ContextValue == nil && p.LiteralValue == nil && p.Usage == nil {
		return fmt.Errorf("parameter '%s' must have usage", p.Name)
	}
	if p.ContextValue != nil && p.LiteralValue != nil {
		return fmt.Errorf("parameter '%s' cannot set both context_value and literal_value",
			p.Name)
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
		for column, formatterName := range *r.Formatters {
			if err := ColumnFormatterType(formatterName).validate(); err != nil {
				return fmt.Errorf("unknown formatter for column %s: %w", column, err)
			}
		}
	}
	if r.ComputedColumns != nil {
		for column, columnFunction := range *r.ComputedColumns {
			if err := ColumnFunctionType(columnFunction).validate(); err != nil {
				return fmt.Errorf("unknown computed column %s: %w", column, err)
			}
		}
	}
	if r.LocalAction != nil {
		err := LocalActionType(*r.LocalAction).validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ResultModel) GetFormatter(column string) print.Formatter {
	if r.Formatters != nil {
		return ColumnFormatterType((*r.Formatters)[column]).GetFormatter()
	}
	return nil
}

func (r *ResultModel) GetColumnFunction(column string) print.ColumnFunction {
	if r.ComputedColumns != nil {
		f, _ := ColumnFunctionType((*r.ComputedColumns)[column]).GetColumnFunction()
		return f
	}
	return nil
}

func warnIfTruncated(result *jnode.Node, truncationIndicatorPath []string) error {
	n := print.Nav(result, truncationIndicatorPath)
	if n.AsBool() {
		return fmt.Errorf("the server indicated the results were truncated")
	}
	return nil
}
