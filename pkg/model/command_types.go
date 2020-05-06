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
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type CommandType string

var validCommandTypes = []string{
	"group", "print_cluster", "print_client", "cluster", "client",
}

func (t CommandType) validate() error {
	if !util.StringSliceContains(validCommandTypes, string(t)) {
		return fmt.Errorf("invalid type %s, must be one of %s", t, strings.Join(validCommandTypes, " "))
	}
	return nil
}

func (t CommandType) IsGroup() bool {
	return t == "group"
}

func (t CommandType) createOptions() options.Interface {
	switch t {
	case "group":
		return nil
	case "print_cluster":
		return &options.PrintClusterOpts{}
	case "print_client":
		return &options.PrintClientOpts{}
	case "cluster":
		return &options.ClusterOpts{}
	case "client":
		return &options.ClientOpts{}
	}
	return nil
}
