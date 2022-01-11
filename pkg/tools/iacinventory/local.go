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

package iacinventory

import (
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Local struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Single = &Local{}

func (t *Local) Name() string {
	return "local-inventory"
}

func (t *Local) Register(cmd *cobra.Command) {
	t.Internal = true
	t.DirectoryBasedToolOpts.Register(cmd)
}

func (t *Local) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "local",
		Short: "Inventory a directory for infrastructure-as-code",
	}
}

func (t *Local) Run() (*tools.Result, error) {
	log.Infof("Finding local infrastructure-as-code inventory under {primary:%s}", t.GetDirectory())
	m := t.GetInventory()
	n, _ := print.ToResult(m)
	r := &tools.Result{
		Data: n,
	}
	return r, nil
}
