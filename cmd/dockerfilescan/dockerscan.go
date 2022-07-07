// Copyright 2022 Lacework Inc
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

package dockerfilescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/checkov"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&checkov.Dockerfile{})
	c.Use = "dockerfile-scan"
	c.Short = "Scan Dockerfiles"
	c.Long = `Scan Dockerfiles

Use the sub-commands to explicitly choose a scanner to use.`
	ckv := tools.CreateCommand(&checkov.Dockerfile{})
	ckv.Use = "checkov"
	ckv.Short = "Scan Dockerfiles with checkov"
	c.AddCommand(ckv)
	return c
}
