package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

type Factory interface {
	EC2() (EC2, error)
}

func CreateFactory(accessKeyID string, secretAccessKey string, region string) Factory {
	return &factoryImpl{
		accessKeyId:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region: &region,
	}
}

type factoryImpl struct {
	accessKeyId     string
	secretAccessKey string
	region * string
}

func (f *factoryImpl) EC2() (EC2, error) {
	s, err := f.session()

	if err != nil {
		return nil, err
	}

	return ec2.New(s), nil
}

func (f *factoryImpl) session() (*session.Session, error) {
	s, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			f.accessKeyId,
			f.secretAccessKey,
			"",
		),
		Region: f.region,
	})

	if err != nil {
		return nil, err
	}

	return s, errors.Wrapf(err, "failed to create session")
}
