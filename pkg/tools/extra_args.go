package tools

import "github.com/spf13/cobra"

// ExtraArgs captures extra arguments to a command
type ExtraArgs []string

func (ex *ExtraArgs) ArgsValue() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		*ex = ExtraArgs(args)
		return nil
	}
}
