package cmd

import (
	"github.com/knlambert/docker-remote.git/pkg/host"
	"github.com/spf13/cobra"
	"log"
)

func createShellCmd() *cobra.Command {
	downCmd := cobra.Command{
		Use: "shell [path-to-public-key]",
		Short: "Open a shell to the remote host",
		Run: func(cmd *cobra.Command, args[]string) {
			if err := host.BuildHostImplementation("ec2").Shell(&args[0]); err != nil {
				log.Fatal(err)
			}
		},
		Args: cobra.MinimumNArgs(1),
	}

	return &downCmd
}