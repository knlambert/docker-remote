package cmd

import (
	"github.com/knlambert/docker-remote.git/pkg/host"
	"github.com/spf13/cobra"
	"log"
)

func createUpCmd() *cobra.Command {
	upCmd := cobra.Command{
		Use:   "up",
		Short: "Create the docker host.",
		Long:  "Create the docker host.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := host.BuildHostImplementation("ec2").Up(); err != nil {
				log.Fatal(err)
			}
		},
	}

	return &upCmd
}
