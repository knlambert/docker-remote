package host

import (
	"github.com/knlambert/docker-remote.git/pkg/docker"
	"github.com/knlambert/docker-remote.git/pkg/sshutil"
	"github.com/knlambert/docker-remote.git/pkg/std/user"
	"github.com/pkg/errors"
)

type PluginHelpers interface {
	DefaultMetadata() (map[string]string, error)
	RegisterToDocker(name string, dockerHost string) error
	SSHUtils() sshutil.SSHUtils
}

func CreatePluginHelpers() PluginHelpers {
	return &pluginHelperImpl{
		user:   user.CreateUser(),
		docker: docker.CreateDocker(),
		sshUtils: sshutil.CreateSSHUtils(),
	}
}

type pluginHelperImpl struct {
	user   user.User
	docker docker.Docker
	sshUtils sshutil.SSHUtils
}

func (b *pluginHelperImpl) DefaultMetadata() (map[string]string, error) {
	var metadata = map[string]string{}

	currentUser, err := b.user.Current()

	if err != nil {
		return nil, err
	}

	metadata["owner"] = currentUser.Name
	metadata["managed_by"] = "docker-remote"

	return metadata, nil
}

//Registers a docker service on the local machine leveraging the Docker contexts.
func (b *pluginHelperImpl) RegisterToDocker(name string, dockerHost string) error {
	if err := b.docker.ContextSet(name, dockerHost); err != nil {
		return errors.Wrap(err, "failed to save the docker host to a docker context")
	}
	return nil
}

func (b *pluginHelperImpl) SSHUtils() sshutil.SSHUtils {
	return b.sshUtils
}


