package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:   "docker-remote",
	Long: "A command line designed to host a docker remote host in one command",
	Args: cobra.MinimumNArgs(1),
}


func Execute() {

	for _, requestedDriver := range []string{"ec2"} {
		driverCmd := cobra.Command{
			Use:   requestedDriver,
			Short: fmt.Sprintf("%s implementation", requestedDriver),
			Long: fmt.Sprintf("%s implementation", requestedDriver),
			Args: cobra.MinimumNArgs(1),
		}

		rootCmd.AddCommand(&driverCmd)

		driverCmd.AddCommand(createUpCmd(requestedDriver))
		driverCmd.AddCommand(createDownCmd())
		driverCmd.AddCommand(createShellCmd(requestedDriver))

	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
