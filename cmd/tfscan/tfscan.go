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

package tfscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/soluble-ai/soluble-cli/pkg/tools/opal"
	"github.com/soluble-ai/soluble-cli/pkg/tools/terrascan"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfscore"
	"github.com/soluble-ai/soluble-cli/pkg/tools/tfsec"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&checkov.Tool{
		Framework: "terraform",
	})
	c.Use = "terraform-scan"
	c.Aliases = []string{"tf-scan"}
	c.Short = `Scan terraform`
	c.Long = `Scan terraform

Scans terraform code with checkov.  Use a sub-command to explicitly choose a scanner.`
	scan := tools.CreateCommand(&tfscore.Tool{})
	scan.Use = "plan"
	scan.Deprecated = "Use 'terraform-plan scan' instead"
	c.AddCommand(
		tools.CreateCommand(&tfsec.Tool{}),
		tools.CreateCommand(&terrascan.Tool{}),
		tools.CreateCommand(&checkov.Tool{
			Framework: "terraform",
		}),
		scan,
	)
	opal := tools.CreateCommand(&opal.Tool{})
	opal.Hidden = true
	c.AddCommand(opal)
	return c
}
