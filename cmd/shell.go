package cmd

import (
	"github.com/knlambert/docker-remote.git/pkg/host"
	"github.com/spf13/cobra"
	"log"
)

func createShellCmd(requestedDriver string) *cobra.Command {
	impl := host.BuildHostImplementation(requestedDriver)

	shellParams := impl.RegisterCommandParams("shell")

	shellCmd := cobra.Command{
		Use: "shell",
		Short: "Open a shell to the remote host",
		Run: func(cmd *cobra.Command, args[]string) {
			if err := impl.Shell(shellParams); err != nil {
				log.Fatal(err)
			}
		},
	}

	if err := impl.RegisterCobraFlags(&shellCmd, shellParams); err != nil {
		log.Fatal(err)
	}

	return &shellCmd
}