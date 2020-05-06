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

import "github.com/spf13/cobra"

func AddPreRunE(c *cobra.Command, f func(*cobra.Command, []string) error) {
	next := c.PreRunE
	c.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := f(cmd, args); err != nil {
			return err
		}
		if next != nil {
			return next(cmd, args)
		}
		return nil
	}
}
