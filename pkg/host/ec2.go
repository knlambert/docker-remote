package host

import (
	"github.com/knlambert/docker-remote.git/pkg/sdk/aws"
	"log"
	"time"
)

func CreateEC2Host(
	accessKeyID string, secretAccessKey string, region string,
) DockerHostSystem {
	return &ec2HostImpl{
		aws: aws.Create(
			accessKeyID, secretAccessKey, region,
		),
	}
}

type ec2HostImpl struct {
	aws aws.AWS
}

func (e *ec2HostImpl) Up() error {
	instanceId, err := e.aws.CreateVM()

	if err != nil {
		return err
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
	return nil
}
