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
	"os/exec"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/model"
)

func getGCPIdentityToken(name string) (string, error) {
	c := exec.Command("gcloud", "auth", "list", "--filter=status:ACTIVE")
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	account := string(out)
	if strings.Contains(account, "gserviceaccount") {
		c = exec.Command("gcloud", "auth", "print-identity-token", "--token-format=full")
	} else {
		c = exec.Command("gcloud", "auth", "print-identity-token")
	}
	out, err = c.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func init() {
	model.AddContextValueSupplier("gcpIdentityToken", getGCPIdentityToken)
}
