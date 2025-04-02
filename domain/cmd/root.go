package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "rcmetering",
	Short: "redcardinal metering service",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
