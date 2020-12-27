package host

import (
	"fmt"
	"github.com/knlambert/docker-remote.git/pkg/host/aws"
	"github.com/pkg/errors"
	"log"
	"time"
)

func CreateEC2Host(
	accessKeyID string,
	secretAccessKey string,
	region string,
) DockerHostSystem {
	return &ec2HostImpl{
		aws: aws.Create(
			accessKeyID, secretAccessKey, region,
		),
		helpers: CreatePluginHelpers(),
	}
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

func (e *ec2HostImpl) Up() error {
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

func (e *ec2HostImpl) Shell(publicKeyPath *string) error {
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
		*publicKeyPath,
	); err != nil {
		return err
	}

	return nil
}
