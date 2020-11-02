package cmd

import (
	"github.com/knlambert/docker-remote.git/pkg/host"
	"github.com/spf13/cobra"
	"log"
)

func createDownCmd() *cobra.Command {
	downCmd := cobra.Command{
		Use: "down",
		Short: "Cleanup a docker host.",
		Run: func(cmd *cobra.Command, args[]string) {
			if err := host.BuildHostImplementation("ec2").Down(); err != nil {
				log.Fatal(err)
			}
		},
	}

	return &downCmd
}