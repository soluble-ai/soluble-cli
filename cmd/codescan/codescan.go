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

package codescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/bandit"
	"github.com/soluble-ai/soluble-cli/pkg/tools/brakeman"
	"github.com/soluble-ai/soluble-cli/pkg/tools/gosec"
	"github.com/soluble-ai/soluble-cli/pkg/tools/semgrep"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "code-scan",
		Short: "Scan code with a variety of static analysis tools",
	}
	c.AddCommand(
		tools.CreateCommand(&semgrep.Tool{}),
		tools.CreateCommand(&bandit.Tool{}),
		tools.CreateCommand(&brakeman.Tool{}),
		tools.CreateCommand(&gosec.Tool{}),
	)
	return c
}
