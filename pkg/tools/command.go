// Copyright 2021 Soluble Inc
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

package tools

import (
	"fmt"

	"github.com/soluble-ai/soluble-cli/pkg/blurb"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/spf13/cobra"
)

type HasCommandTemplate interface {
	CommandTemplate() *cobra.Command
}

func CreateCommand(tool Interface) *cobra.Command {
	var c *cobra.Command
	if ct, ok := tool.(HasCommandTemplate); ok {
		c = ct.CommandTemplate()
		if c.Args == nil {
			c.Args = cobra.NoArgs
		}
	} else {
		c = &cobra.Command{
			Use:   tool.Name(),
			Short: fmt.Sprintf("Run %s", tool.Name()),
			Args:  cobra.NoArgs,
		}
	}
	c.RunE = func(cmd *cobra.Command, args []string) error {
		return runTool(tool)
	}
	tool.Register(c)
	return c
}

func runTool(tool Interface) error {
	opts := tool.GetToolOptions()
	opts.Tool = tool
	result, err := opts.RunTool(true)
	if err != nil || result == nil {
		return err
	}
	if !opts.UploadEnabled {
		blurb.SignupBlurb(opts, "Want to manage findings with {primary:Soluble}?", "run this command again with the {info:--upload} flag")
	}
	if result.Assessment != nil && result.Assessment.URL != "" {
		log.Infof("Results uploaded, see {primary:%s} for more information", result.Assessment.URL)
	}
	return nil
}
