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

package cfnscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	cfnpythonlint "github.com/soluble-ai/soluble-cli/pkg/tools/cfn-python-lint"
	"github.com/soluble-ai/soluble-cli/pkg/tools/cfnnag"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&cfnpythonlint.Tool{})
	c.Use = "cloudformation-scan"
	c.Aliases = []string{"cfn-scan"}
	c.Short = "Scan cloudformation templates"
	c.Long = `Scan cloudformation templates with cfn-python-lint by default.

Use the sub-commands to explicitly choose a scanner to use.`
	c.AddCommand(
		tools.CreateCommand(&cfnnag.Tool{}),
		tools.CreateCommand(&cfnpythonlint.Tool{}),
		tools.CreateCommand(&checkov.Tool{
			Framework: "cloudformation",
		}),
	)
	return c
}
