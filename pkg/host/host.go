package host

import "github.com/spf13/cobra"

type DockerHostSystem interface {
	Down() error
	Up(params interface{}) error
	RegisterCommandParams(command string) interface{}
	RegisterCobraFlags(
		cmd *cobra.Command, upParams interface{},
	) error
	Shell(params interface{}) error
}
