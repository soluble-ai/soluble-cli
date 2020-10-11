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

package scan

import (
	"github.com/spf13/cobra"
)

var (
	// PolicyPath Policy path directory
	PolicyPath string

	// IacFilePath Path to a single IaC file
	IacFilePath string

	// IacDirPath Path to a directory containing one or more IaC files
	IacDirPath string

	// ConfigOnly will output resource config (should only be used for debugging purposes)
	ConfigOnly bool
)

// Command for scan that will be registered with the root command of cli
func Command() *cobra.Command {
	scanCmd := &cobra.Command{
		Use:   "scan [flags]",
		Short: "scans the terraform code for config errors and vulnerabilities",
		Long:  `soluble scan is a simple tool to detect potential security vulnerabilities in your terraform based infrastructure code.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	scanCmd.Flags().StringVarP(&IacFilePath, "iac-file", "f", "", "path to a single IaC file")
	scanCmd.Flags().StringVarP(&IacDirPath, "iac-dir", "d", ".", "path to a directory containing one or more IaC files")
	scanCmd.Flags().StringVarP(&PolicyPath, "policy-path", "p", "", "policy path directory")
	return scanCmd
}
