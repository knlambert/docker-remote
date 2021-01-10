package cmd

import (
	"github.com/knlambert/docker-remote.git/pkg/host"
	"github.com/spf13/cobra"
)

func createUpCmd(requestedDriver string) *cobra.Command {
	impl := host.BuildHostImplementation(requestedDriver)
	return impl.CobraCommand(host.Up)
}
