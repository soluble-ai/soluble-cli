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
	"embed"
	"testing"
)

//go:embed testdata
var testModelFS embed.FS

func TestModel(t *testing.T) {
	source := &FileSystemSource{
		Filesystem: testModelFS,
	}
	err := Load(source)
	if err != nil {
		t.Error(err)
	}
	if len(Models) != 1 || len(Models[0].Command.Commands) != 1 {
		t.Error(Models)
	}
	pingModel := Models[0].Command.Commands[0]
	if pingModel.Name != "ping" {
		t.Fatal(pingModel)
	}
	ping := pingModel.GetCommand()
	f := ping.GetCobraCommand().Flag("dummy-id")
	if f == nil || f.Usage != "dummy value" {
		t.Error(f)
	}
	_ = f.Value.Set("1")
	_ = ping.GetCobraCommand().Flag("action").Value.Set("update")
	context := NewContextValues()
	params, _, err := pingModel.processParameters(ping.GetCobraCommand(), context)
	if err != nil {
		t.Fatal(err)
	}
	if params["action"] != "update" {
		t.Error(params)
	}
	if val, err := context.Get("dummyID"); err != nil || val != "1" {
		t.Error(err, val)
	}
	if path := pingModel.getPath(context); path != "ping/1" {
		t.Error(path)
	}
}
