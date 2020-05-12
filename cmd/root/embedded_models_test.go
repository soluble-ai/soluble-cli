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

package root

import (
	"strings"
	"testing"

	"github.com/soluble-ai/soluble-cli/pkg/model"
)

func TestModel(t *testing.T) {
	err := model.Load(getEmbeddedModelsSource())
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Models) < 5 {
		t.Error("wrong # of models found")
	}
	var orgModel *model.Model
	for _, model := range model.Models {
		if strings.HasSuffix(model.FileName, "/org.hcl") {
			if model.APIPrefix != "/api/v1" {
				t.Error(model)
			}
			orgModel = model
			break
		}
	}
	if orgModel == nil {
		t.Fatal("can't find org model")
	}
	if orgModel.Command.Type != "group" || !orgModel.Command.GetCommandType().IsGroup() {
		t.Error()
	}
	_ = orgModel.Command.GetCommand()
}
