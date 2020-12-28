package host

import (
	"fmt"
	"github.com/knlambert/docker-remote.git/pkg/host/aws"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
	"time"
)

func CreateEC2Host() DockerHostSystem {
	return &ec2HostImpl{
		aws:     aws.Create(),
		helpers: CreatePluginHelpers(),
	}
}

type UpParams struct {
	AMI           string
	InstanceType  string
	KeyName       string
	SecurityGroup string
}

type ShellParams struct {
	PathToPublicKey string
}

type ec2HostImpl struct {
	aws     aws.AWS
	helpers PluginHelpers
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

func (e *ec2HostImpl) RegisterCobraFlags(
	cmd *cobra.Command,
	commandParams interface{},
) error {
	switch casted := commandParams.(type) {
	case *ShellParams:
		cmd.Flags().StringVarP(
			&casted.PathToPublicKey,
			"public-key-path", "", "",
			"The path to the PEM key file to use for connection",
		)
		_ = cmd.MarkFlagRequired("public-key-path")
	case *UpParams:
		cmd.Flags().StringVarP(
			&casted.AMI, "ami", "", "ami-0c2f25c1f66a1ff4d", "The AMI to use",
		)

		cmd.Flags().StringVarP(
			&casted.InstanceType, "instance-type", "", "t2.micro", "The instance type to use",
		)

		cmd.Flags().StringVarP(
			&casted.KeyName, "key-name", "", "", "The ssh key pair to use to connect to the VM",
		)

		cmd.Flags().StringVarP(
			&casted.SecurityGroup, "sg-id", "", "", "A security group ID for the VM",
		)

		_ = cmd.MarkFlagRequired("key-name")
		_ = cmd.MarkFlagRequired("sg-id")

	}

	return nil
}

func (e *ec2HostImpl) RegisterCommandParams(command string) interface{} {
	switch command {
	case "shell":
		return &ShellParams{}
	case "up":
		return &UpParams{}

	}
	return nil
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

	if err := e.helpers.SSHConnection(
		*instance.PublicIp,
		"ec2-user",
		shellParams.PathToPublicKey,
	); err != nil {
		return err
	}

	return nil
}
