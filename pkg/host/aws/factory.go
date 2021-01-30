package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

type Factory interface {
	EC2() (EC2, error)
}

func CreateFactory() Factory {
	return &factoryImpl{}
}

type factoryImpl struct {}

func (f *factoryImpl) EC2() (EC2, error) {
	s, err := f.session()

	if err != nil {
		return nil, err
	}

	return ec2.New(s), nil
}

func (f *factoryImpl) session() (*session.Session, error) {
	s, err := session.NewSession()


	if err != nil {
		return nil, err
	}

	return s, errors.Wrapf(err, "failed to create session")
}
