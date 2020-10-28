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

package query

import (
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "query",
		Short: "Commands list and run Soluble queries",
	}
	c.AddCommand(listCommand())
	c.AddCommand(listParametersCommand())
	c.AddCommand(runCommand())
	return c
}

func listCommand() *cobra.Command {
	opts := &options.PrintClientOpts{
		PrintOpts: options.PrintOpts{
			Path:    []string{"data"},
			Columns: []string{"name", "description"},
		},
	}
	c := &cobra.Command{
		Use:   "list",
		Short: "List available queries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := opts.GetAPIClient().Get("/api/v1/org/{org}/queries")
			if err != nil {
				return err
			}
			opts.PrintResult(result)
			return nil
		},
	}
	opts.Register(c)
	return c
}

func listParametersCommand() *cobra.Command {
	opts := &options.PrintClientOpts{
		PrintOpts: options.PrintOpts{
			Path:    []string{"parameters"},
			Columns: []string{"name", "required", "description"},
		},
	}
	var queryName string
	c := &cobra.Command{
		Use:   "list-parameters",
		Short: "List the parameters of a query",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := opts.GetAPIClient().Get("/api/v1/org/{org}/queries")
			if err != nil {
				return err
			}
			for _, query := range result.Path("data").Elements() {
				if query.Path("name").AsText() == queryName {
					opts.PrintResult(query)
					return nil
				}
			}
			return fmt.Errorf("cannot find query %s", queryName)
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&queryName, "query-name", "", "The query name")
	_ = c.MarkFlagRequired("query-name")
	return c
}

func runCommand() *cobra.Command {
	opts := &options.PrintClientOpts{
		PrintOpts: options.PrintOpts{
			Path:        []string{"data"},
			WideColumns: []string{},
		},
	}
	var queryName string
	var parameters map[string]string
	var textSearch string
	c := &cobra.Command{
		Use:   "run",
		Short: "Run a query",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := fmt.Sprintf("/api/v1/org/{org}/queries/%s", queryName)
			if textSearch != "" {
				parameters["q"] = textSearch
			}
			result, err := opts.GetAPIClient().GetWithParams(path, parameters)
			if err != nil {
				return err
			}
			for _, field := range result.Path("metadata").Path("fields").Elements() {
				name := field.Path("name").AsText()
				opts.Columns = append(opts.Columns, name)
				if hasDisplayHint(field, "WIDE") {
					opts.WideColumns = append(opts.WideColumns, name)
				}
				if hasDisplayHint(field, "TS") {
					opts.SetFormatter(name, print.TimestampFormatter)
				} else if hasDisplayHint(field, "RELATIVE_TS") {
					opts.SetFormatter(name, print.RelativeTimestampFormatter)
				}
			}
			opts.PrintResult(result)
			return nil
		},
	}
	opts.Register(c)
	c.Flags().StringVar(&queryName, "query-name", "", "The name of the query to run")
	c.Flags().StringToStringVarP(&parameters, "parameters", "p", map[string]string{}, "Parameter values in the form name=value")
	c.Flags().StringVar(&textSearch, "text", "", "Search against \"interesting\" fields, equivalent to '-p q=text'")
	_ = c.MarkFlagRequired("query-name")
	return c
}

func hasDisplayHint(field *jnode.Node, hint string) bool {
	for _, n := range field.Path("displayHints").Elements() {
		if n.AsText() == hint {
			return true
		}
	}
	return false
}
