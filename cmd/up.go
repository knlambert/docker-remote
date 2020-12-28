package cmd

import (
	"github.com/knlambert/docker-remote.git/pkg/host"
	"github.com/spf13/cobra"
	"log"
)

func createUpCmd(requestedDriver string) *cobra.Command {
	impl := host.BuildHostImplementation(requestedDriver)

	upParams := impl.RegisterCommandParams("up")

	upCmd := cobra.Command{
		Use:   "up",
		Short: "Creates the docker host.",
		Long:  "Creates the docker host.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := impl.Up(upParams); err != nil {
				log.Fatal(err)
			}
		},
	}

	if err := impl.RegisterCobraFlags(&upCmd, upParams); err != nil {
		log.Fatal(err)
	}

	return &upCmd
}
