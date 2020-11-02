package user

import "os/user"

type User interface {
	Current()  (*user.User, error)
}

func CreateUser() User {
	return &userImpl{}
}

type userImpl struct {}

func (u *userImpl) Current() (*user.User, error) {
	return user.Current()
}
