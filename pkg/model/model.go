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

func (m *Model) validate() error {
	return m.Command.validate(m)
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

func (cm *CommandModel) validate(m *Model) error {
	cm.model = m
	if cm.Use != nil && strings.HasPrefix(*cm.Use, cm.Name+" ") {
		return fmt.Errorf("use must begin with with name %s", cm.Name)
	}
	if err := cm.GetCommandType().validate(); err != nil {
		return fmt.Errorf("invalid command type for %s: %w", cm.Name, err)
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
			contextValues.values[p.Name] = value
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
		for column, formatterName := range *r.Formatters {
			if err := ColumnFormatterType(formatterName).validate(); err != nil {
				return fmt.Errorf("invalid formatter for column %s: %w", column, err)
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
