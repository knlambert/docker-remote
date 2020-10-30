package main

import (
	"github.com/knlambert/docker-remote.git/pkg/host"
	"log"
	"os"
)

func main() {
	var region = "ca-central-1"

	if os.Getenv("AWS_DEFAULT_REGION") != "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}

	ec2 := host.CreateEC2Host(
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		region,
	)

	if err := ec2.Up(); err != nil {
		log.Fatal(err)
	}
}
