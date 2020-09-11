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
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/model"
)

func isDefaultClusterID(clusterID string) bool {
	return clusterID == config.Config.DefaultClusterID
}

func computeIsDefaultCluster(n *jnode.Node) interface{} {
	if isDefaultClusterID(n.Path("clusterId").AsText()) {
		return "   *"
	}
	return ""
}

func init() {
	model.RegisterColumnFunction("is_default_cluster", computeIsDefaultCluster)
}
