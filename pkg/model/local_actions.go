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
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/options"
)

type LocalAction string

type Action func(opts options.Interface, n *jnode.Node) (*jnode.Node, error)

var actions = map[string]Action{}

func (a LocalAction) validate() error {
	if a == "" || actions[string(a)] != nil {
		return nil
	}
	return fmt.Errorf("unknown local_action %s", a)
}

func (a LocalAction) Run(opts options.Interface, n *jnode.Node) (*jnode.Node, error) {
	return actions[string(a)](opts, n)
}

func RegisterAction(name string, action Action) {
	actions[name] = action
}
