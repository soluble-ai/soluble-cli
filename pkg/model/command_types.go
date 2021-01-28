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
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"

	"github.com/spf13/cobra"
)

type CommandType string

type CommandMaker func(c *cobra.Command, cm *CommandModel) Command

type Command interface {
	GetAPIClient() *api.Client
	GetUnauthenticatedAPIClient() *api.Client
	PrintResult(n *jnode.Node)
	GetCobraCommand() *cobra.Command
	SetContextValues(c map[string]string)
}

type GroupCommand struct {
	CobraCommand *cobra.Command
	Commands     []Command
}

var commandTypes = map[string]CommandMaker{}

func RegisterCommandType(name string, creator CommandMaker) {
	commandTypes[name] = creator
}

func (t CommandType) validate() error {
	if _, ok := commandTypes[string(t)]; !ok {
		return fmt.Errorf("invalid type %s", t)
	}
	return nil
}

func (t CommandType) IsGroup() bool {
	return t == "group"
}

func (t CommandType) makeCommand(c *cobra.Command, cm *CommandModel) Command {
	return commandTypes[string(t)](c, cm)
}

func (g *GroupCommand) GetAPIClient() *api.Client {
	return nil
}
func (g *GroupCommand) GetUnauthenticatedAPIClient() *api.Client {
	return nil
}
func (g *GroupCommand) PrintResult(n *jnode.Node) {}
func (g *GroupCommand) GetCobraCommand() *cobra.Command {
	return g.CobraCommand
}
func (g *GroupCommand) SetContextValues(m map[string]string) {}

func init() {
	RegisterCommandType("group", func(c *cobra.Command, cm *CommandModel) Command {
		g := &GroupCommand{
			CobraCommand: c,
		}
		for i := range cm.Commands {
			command := cm.Commands[i].GetCommand()
			g.Commands = append(g.Commands, command)
			c.AddCommand(command.GetCobraCommand())
		}
		return g
	})
}
