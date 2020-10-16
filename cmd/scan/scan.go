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
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/scanner"
	"github.com/spf13/cobra"
)

var (
	// policyPath Policy path directory, if you want to specify your custom policies
	policyPath string

	// iacFilePath Path to a single IaC file
	iacFilePath string

	// iacDirPath Path to a directory containing one or more IaC files
	iacDirPath string

	// report to control plane
	report bool
)

// Command for scan that will be registered with the root command of cli
func Command() *cobra.Command {
	opts := options.PrintOpts{}
	scanCmd := &cobra.Command{
		Use:   "scan [flags]",
		Short: "scans the terraform code for config errors and vulnerabilities",
		Long:  `soluble scan is a simple tool to detect potential compliance and security in the terraform based Infrastructure as Code.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// create a new runtime executor for processing IaC
			scanner, err := scanner.NewScanner(iacFilePath, iacDirPath, policyPath, report)
			if err != nil {
				log.Errorf("Failed to create new scanner %s", err.Error())
				return err
			}
			// scanner output
			results, err := scanner.Execute()
			if err != nil {
				log.Errorf("Failed to create new scanner %s", err.Error())
				return err
			}
			log.Infof("VIOLATIONS:")

			violations := results.Path("results").Path("violations")
			opts.PrintResult(violations)

			log.Infof("Results:")
			opts.PrintResult(results.Path("results").Path("count"))

			// pretty print for demo
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Severity", "Name"})
			for _, v := range violations.Elements() {
				output := []string{v.Path("ruleId").AsText(), v.Path("severity").AsText(), v.Path("ruleName").AsText()}
				table.Append(output)
			}
			table.Render()
			return nil
		},
	}
	opts.Register(scanCmd)
	scanCmd.Flags().StringVarP(&iacFilePath, "iac-file", "f", "", "path to a single IaC file")
	scanCmd.Flags().StringVarP(&iacDirPath, "iac-dir", "d", ".", "path to a directory containing one or more IaC files")
	scanCmd.Flags().StringVarP(&policyPath, "policy-path", "p", "", "policy path directory")
	// scanCmd.Flags().StringVarP(&format, "format", "o", "yaml", "output type (json, yaml, table)")
	scanCmd.Flags().BoolVarP(&report, "report", "r", true, "report back to control plane")
	return scanCmd
}
