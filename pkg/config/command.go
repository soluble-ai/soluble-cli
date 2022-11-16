package config

import "github.com/spf13/cobra"

const ConfigurationNotRequired = "ConfigurationNotRequired"

func IsConfigurationRequired(cmd *cobra.Command) bool {
	if cmd.Run == nil && cmd.RunE == nil {
		// command groups don't require configuration
		return false
	}
	return cmd.Annotations["ConfigurationNotRequired"] == ""
}
