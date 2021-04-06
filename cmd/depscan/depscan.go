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

package depscan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
<<<<<<< HEAD
	bundleraudit "github.com/soluble-ai/soluble-cli/pkg/tools/bundler-audit"
=======
	"github.com/soluble-ai/soluble-cli/pkg/tools/npmaudit"
>>>>>>> master
	"github.com/soluble-ai/soluble-cli/pkg/tools/retirejs"
	"github.com/soluble-ai/soluble-cli/pkg/tools/trivyfs"
	"github.com/soluble-ai/soluble-cli/pkg/tools/yarnaudit"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&trivyfs.Tool{})
	c.Use = "dep-scan"
	c.Short = "Scan application dependencies"
	c.Long = `Scan application dependencies with trivy by default`
	c.AddCommand(
		tools.CreateCommand(&trivyfs.Tool{}),
		tools.CreateCommand(&retirejs.Tool{}),
		tools.CreateCommand(&bundleraudit.Tool{}),
		tools.CreateCommand(&npmaudit.Tool{}),
		tools.CreateCommand(&yarnaudit.Tool{}),
	)
	return c
}
