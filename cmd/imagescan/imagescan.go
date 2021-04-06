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

package imagescan

import (
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/tools/hadolint"
	"github.com/soluble-ai/soluble-cli/pkg/tools/trivy"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := tools.CreateCommand(&trivy.Tool{})
	c.Use = "image-scan"
	c.Short = "Scan a container image"
	c.Long = `Scan a container image for vulnerabilities of OS packages or lint a container image for best practices`
	c.AddCommand(
		tools.CreateCommand(&trivy.Tool{}),
		tools.CreateCommand(&hadolint.Tool{}),
	)
	return c
}
