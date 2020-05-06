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

package options

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func TestAddPreRun(t *testing.T) {
	ac := 0
	a := func(cmd *cobra.Command, args []string) error {
		ac += 1
		return nil
	}
	bc := 0
	b := func(cmd *cobra.Command, args []string) error {
		bc += 1
		if bc == 1 {
			return nil
		}
		return fmt.Errorf("2nd time")
	}
	c := &cobra.Command{
		PreRunE: a,
	}
	AddPreRunE(c, b)
	err := c.PreRunE(c, nil)
	if err != nil {
		t.Error(err)
	}
	if ac != 1 || bc != 1 {
		t.Error()
	}
	err = c.PreRunE(c, nil)
	if ac != 1 || bc != 2 || err == nil {
		t.Error(ac, bc, err)
	}
}
