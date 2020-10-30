package aws

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

var initScript string = `#!/bin/bash
sudo yum update -y
sudo amazon-linux-extras install docker -y
sudo yum install docker -y
sudo service docker start
sudo usermod -a -G docker ec2-user
`


type AWS interface {
	CreateVM() (*string, error)
	InstanceIsReady(instanceId string) (bool, error)
}

func Create(accessKeyID string, secretAccessKey string, region string) AWS {
	return &awsImpl{
		CreateFactory(
			accessKeyID,
			secretAccessKey,
			region,
		),
	}
}

type awsImpl struct {
	factory Factory
}

func (a *awsImpl) CreateVM() (*string, error) {
	c, err := a.factory.EC2()

	if err != nil {
		return nil, err
	}

	res, err := c.RunInstances(&ec2.RunInstancesInput{
		ImageId:                           makeStringPnt("ami-0c2f25c1f66a1ff4d"),
		InstanceType:                      makeStringPnt("t2.micro"),
		MaxCount:makeInt64Pnt(1),
		MinCount: makeInt64Pnt(1),
		UserData: makeStringPnt(base64.StdEncoding.EncodeToString([]byte(initScript))),
		SecurityGroupIds: []*string{makeStringPnt("sg-0aa7c79fa8f29db48")},
		KeyName: makeStringPnt("ec2-admin"),
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to kick a VM in EC2")
	}

	return res.Instances[0].InstanceId, nil
}

func (a *awsImpl) InstanceIsReady(instanceId string) (bool, error) {
	c, err := a.factory.EC2()

	if err != nil {
		return false, err
	}

	res, err := c.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceId},
	})

	if err != nil {
		return false, errors.Wrapf(err, "can't describe instance %s status", instanceId)
	}

	return *res.Reservations[0].Instances[0].State.Name == "running", nil
}

func makeStringPnt(val string) *string {
	return &val
}

func makeInt64Pnt(val int64) *int64 {
	return &val
}