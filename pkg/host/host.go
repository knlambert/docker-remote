package host

import "github.com/spf13/cobra"

type Command string

const (
	Down        Command = "down"
	PortForward Command = "port-forward"
	Shell       Command = "shell"
	Up          Command = "up"
)

type DockerHostSystem interface {
	CobraCommand(
		command Command,
	) *cobra.Command
	Down() error
	PortForward(params interface{}) error
	Shell(params interface{}) error
	Up(params interface{}) error
}
