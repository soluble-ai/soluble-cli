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

package login

import (
	"fmt"
	"os"
)

type HeadlessLeg struct{}

var _ AuthCodeLeg = &HeadlessLeg{}

func (h *HeadlessLeg) GetCode(appURL, state string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/auth/cli-login/%s/0", appURL, state)
	fmt.Printf(`To complete the login process, open the following URL in your browser:

  %s

Paste the resulting authorization code here: `, url)
	_ = os.Stdout.Sync()
	var code string
	_, err := fmt.Scanln(&code)
	return code, err
}
