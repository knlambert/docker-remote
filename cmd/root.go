package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:   "docker-remote",
	Long: "A command line designed to host a docker remote host in one command",
	Args: cobra.MinimumNArgs(1),
}

func Execute() {

	rootCmd.AddCommand(createUpCmd())
	rootCmd.AddCommand(createDownCmd())
	rootCmd.AddCommand(createShellCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
