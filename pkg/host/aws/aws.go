package aws

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
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
	InstanceCreate(
		ami string,
		instanceType string,
		keyName string,
		securityGroup string,
		tags map[string]string,
	) (*string, error)
	InstanceDescribe(
		tags map[string]string,
		states []string,
	) (*InstanceDescription, error)
	InstanceIsReady(instanceId string) (bool, error)
	InstanceTerminate(instanceId string) error
}

func Create() AWS {
	return &awsImpl{
		CreateFactory(),
	}
}

type awsImpl struct {
	factory Factory
}

func (a *awsImpl) InstanceCreate(
	ami string,
	instanceType string,
	keyName string,
	securityGroup string,
	tags map[string]string,
) (*string, error) {
	c, err := a.factory.EC2()

	if err != nil {
		return nil, err
	}

	res, err := c.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(ami),
		InstanceType:     aws.String(instanceType),
		MaxCount:         aws.Int64(1),
		MinCount:         aws.Int64(1),
		UserData:         aws.String(base64.StdEncoding.EncodeToString([]byte(initScript))),
		SecurityGroupIds: []*string{aws.String(securityGroup)},
		KeyName:          aws.String(keyName),
		TagSpecifications: []*ec2.TagSpecification{{
			ResourceType: aws.String("instance"),
			Tags:         mapToTags(tags),
		}},
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to kick a VM in EC2")
	}

	return res.Instances[0].InstanceId, nil
}

type InstanceDescription struct {
	Id *string
	PublicIp *string
}

func (a *awsImpl) InstanceDescribe(
	tags map[string]string, states []string,
) (*InstanceDescription, error) {
	c, err := a.factory.EC2()

	if err != nil {
		return nil, err
	}

	var filters = mapToTagFilter(tags)

	res, err := c.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: filters,
	})

	if err != nil {
		return nil, err
	}

	var instances []*ec2.Instance

	for r := range res.Reservations {

		for i := range res.Reservations[r].Instances {
			instanceState := *res.Reservations[r].Instances[i].State.Name

			if sliceContainsString(states, instanceState) {
				instances = append(instances, res.Reservations[r].Instances[i])
			}
		}
	}

	if len(instances) > 0 {
		return &InstanceDescription{
			Id: instances[0].InstanceId,
			PublicIp: instances[0].PublicIpAddress,
		}, nil
	}

	return nil, nil

}

func (a *awsImpl) InstanceIsReady(instanceId string) (bool, error) {
	c, err := a.factory.EC2()

	if err != nil {
		return false, err
	}

	res, err := c.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	})

	if err != nil {
		return false, errors.Wrap(err, "can't describe instance status")
	}

	return *res.Reservations[0].Instances[0].State.Name == "running", nil
}

func (a *awsImpl) InstanceTerminate(instanceId string) error {
	c, err := a.factory.EC2()

	if err != nil {
		return err
	}

	_, err = c.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{&instanceId},
	})

	if err != nil {
		return err
	}

	return nil
}
func mapToTags(t map[string]string) []*ec2.Tag {
	var r []*ec2.Tag

	for key, value := range t {
		r = append(r, &ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	return r
}

func mapToTagFilter(t map[string]string) []*ec2.Filter {
	var r []*ec2.Filter

	for key, value := range t {
		r = append(r, &ec2.Filter{
			Name: aws.String(fmt.Sprintf("tag:%s", key)),
			Values: []*string{
				aws.String(value),
			},
		})
	}
	return r
}

func sliceContainsString(s []string, v string) bool {
	for i := range s {
		if s[i] == v {
			return true
		}
	}
	return false
}
