package host

import (
	"fmt"
	"github.com/knlambert/docker-remote.git/pkg/docker"
	"github.com/knlambert/docker-remote.git/pkg/std/user"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type PluginHelpers interface {
	DefaultMetadata() (map[string]string, error)
	RegisterToDocker(name string, dockerHost string) error
	SSHConnection(
		host string,
		username string,
		publicKeyPath string,
	) error
}

func CreatePluginHelpers() PluginHelpers {
	return &pluginHelperImpl{
		user:   user.CreateUser(),
		docker: docker.CreateDocker(),
	}
}

type pluginHelperImpl struct {
	user   user.User
	docker docker.Docker
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

func (b *pluginHelperImpl) SSHConnection(
	host string,
	username string,
	publicKeyPath string,
) error {
	key, err := publicKey(publicKeyPath)

	if err != nil {
		log.Fatal(err)
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{key},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)

	if err != nil {
		return err
	}

	session, err := conn.NewSession()

	if err != nil {
		return err
	}

	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.ECHOCTL:       0,
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}
	go io.Copy(stdin, os.Stdin)

	stdinFd := int(os.Stdin.Fd())
	if terminal.IsTerminal(stdinFd) {
		originalState, err := terminal.MakeRaw(stdinFd)

		if err != nil {
			return err
		}

		defer terminal.Restore(stdinFd, originalState)

		termWidth, termHeight, err := terminal.GetSize(stdinFd)

		if err != nil {
			return err
		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			return err
		}
	}

	stdout, err := session.StdoutPipe()

	if err != nil {
		log.Fatal(err)
	}
	go io.Copy(os.Stdout, stdout)

	stderr, err := session.StderrPipe()

	if err != nil {
		return err
	}

	go io.Copy(os.Stderr, stderr)

	err = session.Shell()

	if err != nil {
		return err
	}

	session.Wait()

	return nil
}

func publicKey(path string) (ssh.AuthMethod, error) {
	key, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)

	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}
