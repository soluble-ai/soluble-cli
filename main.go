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

package main

//go:generate go run gen_embed.go

import (
	"os"
	"strings"

	"github.com/soluble-ai/go-colorize"
	"github.com/soluble-ai/soluble-cli/cmd/root"
	_ "github.com/soluble-ai/soluble-cli/pkg/tools/buildreport/github"
)

func main() {
	if opts := os.Getenv("SOLUBLE_OPTS"); opts != "" {
		args := []string{os.Args[0]}
		args = append(args, strings.Split(opts, " ")...)
		args = append(args, os.Args[1:]...)
		os.Args = args
	}
	cmd := root.Command()
	if err := cmd.Execute(); err != nil {
		colorize.Colorize("{danger:Error:} {warning:%s}\n", err)
		os.Exit(1)
	}
}
