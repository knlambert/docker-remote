package host

import (
	"fmt"
	"github.com/knlambert/docker-remote.git/pkg/host/aws"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func CreateEC2Host() DockerHostSystem {
	return &ec2HostImpl{
		aws:     aws.Create(),
		helpers: CreatePluginHelpers(),
	}
}

type ec2HostImpl struct {
	aws     aws.AWS
	helpers PluginHelpers
}

type ForwardParams struct {
	LocalPort       uint
	PathToPublicKey string
	RemotePort      uint
	RemoteAddr      string
}

func (e *ec2HostImpl) PortForward(params interface{}) error {
	fwdParams := params.(*ForwardParams)

	metadata, err := e.helpers.DefaultMetadata()

	if err != nil {
		return errors.Wrap(err, "failed to calculate metadata")
	}

	instance, err := e.aws.InstanceDescribe(metadata, []string{"running", "pending"})

	if err != nil {
		return err
	}

	fmt.Printf("Port-forwarding %d -> %d\n", fwdParams.LocalPort, fwdParams.RemotePort)
	return e.helpers.SSHUtils().LocalPortForward(
		fwdParams.LocalPort,
		fwdParams.RemoteAddr,
		fwdParams.RemotePort,
		*instance.PublicIp,
		"ec2-user",
		fwdParams.PathToPublicKey,
	)
}

func (e *ec2HostImpl) Down() error {
	metadata, err := e.helpers.DefaultMetadata()

	if err != nil {
		return errors.Wrap(err, "failed to calculate metadata")
	}

	instance, err := e.aws.InstanceDescribe(metadata, []string{"running", "pending"})

	if err != nil {
		return err
	}

	if instance != nil {
		if err := e.aws.InstanceTerminate(*instance.Id); err != nil {
			return errors.Wrapf(err, "failed to shutdown the docker host")
		}

		log.Println("Shutdown signal sent")
	} else {
		log.Println("Docker host is already down")
	}

	return nil
}

type UpParams struct {
	AMI           string
	InstanceType  string
	KeyName       string
	SecurityGroup string
}

func (e *ec2HostImpl) CobraCommand(
	command Command,
) *cobra.Command {
	switch command {
	case Down:
		return &cobra.Command{
			Use:   string(command),
			Short: "Cleanup a docker host",
			Run: func(cmd *cobra.Command, args []string) {
				if err := e.Down(); err != nil {
					log.Fatal(err)
				}
			},
		}
	case PortForward:
		fwdParams := ForwardParams{}
		fwdCmd := cobra.Command{
			Use:   string(command),
			Args:  cobra.MinimumNArgs(1),
			Short: "Forward the connection from the remote host",
			PreRunE: func(cmd *cobra.Command, args []string) error {
				r := regexp.MustCompile(`(\d{2,5}):(\d{2,5})`)

				if len(r.FindStringSubmatch(args[0])) == 0 {
					return errors.Errorf("parameter must be 'localPort:remotePort, example: '8080:80'")
				}

				return nil
			},
			Run: func(cmd *cobra.Command, args []string) {
				ports := strings.Split(args[0], ":")

				var converted, _ = strconv.ParseUint(ports[0], 10, 0)
				fwdParams.LocalPort = uint(converted)

				converted, _ = strconv.ParseUint(ports[1], 10, 0)
				fwdParams.RemotePort = uint(converted)

				if err := e.PortForward(&fwdParams); err != nil {
					log.Fatal(err)
				}
			},
		}

		fwdCmd.Flags().StringVarP(
			&fwdParams.PathToPublicKey,
			"public-key-path", "", "",
			"The path to the PEM key file to use for connection",
		)

		fwdCmd.Flags().StringVarP(
			&fwdParams.RemoteAddr,
			"remote-addr", "", "127.0.0.1",
			"The address to forward from on the remote machine",
		)

		_ = fwdCmd.MarkFlagRequired("public-key-path")

		return &fwdCmd

	case Shell:
		shellParams := ShellParams{}
		shellCmd := cobra.Command{
			Use:   string(command),
			Short: "Open a shell to the remote host",
			Run: func(cmd *cobra.Command, args []string) {
				if err := e.Shell(&shellParams); err != nil {
					log.Fatal(err)
				}
			},
		}

		shellCmd.Flags().StringVarP(
			&shellParams.PathToPublicKey,
			"public-key-path", "", "",
			"The path to the PEM key file to use for connection",
		)
		_ = shellCmd.MarkFlagRequired("public-key-path")

		return &shellCmd
	case Up:
		upParams := UpParams{}
		upCmd := cobra.Command{
			Use:   string(command),
			Short: "Creates the docker host",
			Long:  "Creates the docker host",
			Run: func(cmd *cobra.Command, args []string) {
				if err := e.Up(&upParams); err != nil {
					log.Fatal(err)
				}
			},
		}

		upCmd.Flags().StringVarP(
			&upParams.AMI, "ami", "", "ami-0c2f25c1f66a1ff4d", "The AMI to use",
		)

		upCmd.Flags().StringVarP(
			&upParams.InstanceType, "instance-type", "", "t2.micro", "The instance type to use",
		)

		upCmd.Flags().StringVarP(
			&upParams.KeyName, "key-name", "", "", "The sshutil key pair to use to connect to the VM",
		)

		upCmd.Flags().StringVarP(
			&upParams.SecurityGroup, "sg-id", "", "", "A security group ID for the VM",
		)

		_ = upCmd.MarkFlagRequired("key-name")
		_ = upCmd.MarkFlagRequired("sg-id")

		return &upCmd
	}

	return nil
}

type ShellParams struct {
	PathToPublicKey string
}

func (e *ec2HostImpl) Shell(params interface{}) error {
	shellParams := params.(*ShellParams)

	metadata, err := e.helpers.DefaultMetadata()

	if err != nil {
		return errors.Wrap(err, "failed to calculate metadata")
	}

	instance, err := e.aws.InstanceDescribe(metadata, []string{"running", "pending"})

	if err != nil {
		return err
	}

	if instance == nil {
		return errors.Errorf("Please create the host first")
	}

	if err := e.helpers.SSHUtils().SSHConnection(
		*instance.PublicIp,
		"ec2-user",
		shellParams.PathToPublicKey,
	); err != nil {
		return err
	}

	return nil
}
func (e *ec2HostImpl) Up(params interface{}) error {
	upParams := params.(*UpParams)

	metadata, err := e.helpers.DefaultMetadata()

	if err != nil {
		return errors.Wrap(err, "failed to calculate metadata")
	}

	instance, err := e.aws.InstanceDescribe(metadata, []string{"running", "pending"})

	if err != nil {
		return errors.Wrap(err, "failed to describe ec2 host")
	}

	var instanceId *string

	if instance == nil {
		instanceId, err = e.aws.InstanceCreate(
			upParams.AMI,
			upParams.InstanceType,
			upParams.KeyName,
			upParams.SecurityGroup,
			metadata,
		)

		if err != nil {
			return err
		}

	} else {
		instanceId = instance.Id
	}

	var ready = false

	for {
		ready, err = e.aws.InstanceIsReady(*instanceId)

		if err != nil {
			return err
		}

		if ready {
			break
		}

		log.Println("Waiting for instance to be ready ...")

		time.Sleep(5 * time.Second)
	}

	log.Println("Instance is ready !")

	instance, err = e.aws.InstanceDescribe(metadata, []string{"running"})

	if err != nil {
		return err
	}

	log.Printf("Instance IP: %s", *instance.PublicIp)

	dockerContextName := "docker-remote-ec2"

	err = e.helpers.RegisterToDocker(
		dockerContextName,
		fmt.Sprintf("ssh://ec2-user@%s", *instance.PublicIp),
	)

	if err == nil {
		log.Printf("Docker context %s set !", dockerContextName)
	}

	return err
}
