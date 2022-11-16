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

package log

import (
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/spf13/pflag"
)

var (
	trace      bool
	debug      bool
	quiet      bool
	forceColor bool
	logStdout  bool
	logStderr  bool
)

func AddFlags(flags *pflag.FlagSet) {
	// the lacework CLI passes LW_LOG and LW_NOCOLOR, so we'll use
	// those as the default
	flags.BoolVar(&trace, "trace", false, "Run with trace logging")
	flags.BoolVar(&debug, "debug", os.Getenv("LW_LOG") == "DEBUG", "Run with debug logging")
	flags.BoolVar(&quiet, "quiet", false, "Run with no logging")
	flags.BoolVar(&color.NoColor, "no-color", os.Getenv("LW_NOCOLOR") == "true", "Disable color output")
	flags.BoolVar(&forceColor, "force-color", false, "Enable color output")
	flags.BoolVar(&logStdout, "log-stdout", false, "Force the CLI to log to stdout")
	flags.BoolVar(&logStderr, "log-stderr", false, "Force the CLI to log to stderr")
}

func Configure() {
	if forceColor {
		color.NoColor = false
	}
	if quiet {
		Level = Error
	}
	if debug {
		Level = Debug
	}
	if trace {
		Level = Trace
	}
	switch {
	case logStdout:
		return
	case os.Getenv("GITHUB_ACTIONS") == "true":
		// github actions doesn't process interleaved stdout/stderr correctly
		// so if we're running there log and stdout is a terminal, then log to stdout
		if isatty.IsTerminal(os.Stdout.Fd()) {
			return
		}
	case logStderr:
	default:
	}
	color.Output = colorable.NewColorableStderr()
	logStartupMessages()
}
