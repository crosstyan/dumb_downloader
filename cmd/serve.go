package cmd

import "github.com/spf13/cobra"

var serve = cobra.Command{
	Use:   "serve",
	Short: "serve a dumb downloader server",
	Args:  cobra.NoArgs,
	Run:   func(cmd *cobra.Command, args []string) {},
}
