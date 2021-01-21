package log

import (
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/spf13/pflag"
)

var (
	debug      bool
	quiet      bool
	forceColor bool
	logStdout  bool
	logStderr  bool
)

func AddFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&debug, "debug", false, "Run with debug logging")
	flags.BoolVar(&quiet, "quiet", false, "Run with no logging")
	flags.BoolVar(&color.NoColor, "no-color", false, "Disable color output")
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
	switch {
	case logStdout:
	case os.Getenv("GITHUB_ACTIONS") == "true":
		// github actions doesn't process interleaved stdout/stderr correctly
		// so if we're running there log to stdout, otherwise log to stder
	case logStderr:
		fallthrough
	default:
		color.Output = colorable.NewColorableStderr()
	}
}
