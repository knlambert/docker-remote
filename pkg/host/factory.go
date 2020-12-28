package host

import (
	"log"
)

type Driver string

const (
	EC2 Driver = "ec2"
)


func BuildHostImplementation(requestedDriver string) DockerHostSystem {
	driver := Driver(requestedDriver)
	switch driver {
	case EC2:
		return CreateEC2Host()
	}
	log.Fatalf("Can't find any implementation '%s'", requestedDriver)
	return nil
}
