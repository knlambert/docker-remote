package host

import (
	"log"
	"os"
)

type Driver string

const (
	EC2 Driver = "ec2"
)


func BuildHostImplementation(requestedDriver string) DockerHostSystem {
	driver := Driver(requestedDriver)
	switch driver {
	case EC2:
		var region = "ca-central-1"

		if os.Getenv("AWS_DEFAULT_REGION") != "" {
			region = os.Getenv("AWS_DEFAULT_REGION")
		}

		accessKeyId :=  os.Getenv("AWS_ACCESS_KEY_ID")

		if accessKeyId == "" {
			log.Fatal("AWS_ACCESS_KEY_ID env var not set")
		}

		secretAccessKey :=  os.Getenv("AWS_SECRET_ACCESS_KEY")

		if secretAccessKey == "" {
			log.Fatal("AWS_SECRET_ACCESS_KEY env var not set")
		}

		return CreateEC2Host(accessKeyId, secretAccessKey, region)
	}
	log.Fatalf("Can't find any implementation '%s'", requestedDriver)
	return nil
}
