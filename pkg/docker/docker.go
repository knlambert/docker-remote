package docker

import (
	"fmt"
	"github.com/knlambert/docker-remote.git/pkg"
	"github.com/knlambert/docker-remote.git/pkg/std/ioutil"
	"github.com/knlambert/docker-remote.git/pkg/std/os"
	"github.com/knlambert/docker-remote.git/pkg/std/runtime"
	"github.com/knlambert/docker-remote.git/pkg/std/user"
	"github.com/pkg/errors"
)

type Docker interface {
	//Sets a docker context for a specific host.
	ContextSet(
		name string,
		dockerHost string,
	) error
}

func CreateDocker() Docker {
	return &dockerImpl{
		io: ioutil.CreateIOUtil(),
		os:      os.CreateOS(),
		runtime: runtime.CreateRuntime(),
		user:    user.CreateUser(),
	}
}

type dockerImpl struct {
	io      ioutil.IOUtil
	os      os.OS
	runtime runtime.Runtime
	user    user.User
}

//Returns the path to the user's docker config folder.
func (d *dockerImpl) dockerConfigFolderPath() (*string, error) {
	currentUser, err := d.user.Current()

	if err != nil {
		return nil, errors.Wrap(err, "failed to determine current user")
	}

	if currentOs := d.runtime.CurrentOS(); currentOs == "linux" || currentOs == "darwin" {
		folderPath := fmt.Sprintf("%s/.docker", currentUser.HomeDir)
		return &folderPath, nil
	}

	return nil, pkg.CreateNotImplementedError("os not supported")
}
