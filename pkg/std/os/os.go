package os

import "os"

//OS interface to wrap the os package.
type OS interface {
	MkdirAll(path string, perm os.FileMode) error
	PathExists(path string) (bool, error)
}

func CreateOS() OS {
	return &osImpl{}
}

type osImpl struct{}

//Creates a path of folder.
func (o *osImpl) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

//Tests if a path exists
func (o *osImpl) PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
