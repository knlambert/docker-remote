package aws

import "github.com/aws/aws-sdk-go/service/ec2"

type EC2 interface{
	RunInstances(input *ec2.RunInstancesInput) (*ec2.Reservation, error)
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	TerminateInstances(input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
}
