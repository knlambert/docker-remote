package ioutil

import (
	"io/ioutil"
	"os"
)

type IOUtil interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

func CreateIOUtil() IOUtil {
	return &ioUtilImpl{}
}

type ioUtilImpl struct {}

func (i * ioUtilImpl) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}
