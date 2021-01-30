package sshutil

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/Microsoft/go-winio"
	"github.com/knlambert/docker-remote.git/pkg/std/runtime"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

const (
	windowsSSHAgentPipe = `\\.\pipe\openssh-ssh-agent`
	dockerRemoteEC2KeyID = "docker-remote-ec2-key"
)

type SSHUtils interface {
	LocalPortForward(
		localPort uint,
		remoteAddr string,
		remotePort uint,
		host string,
		username string,
	) error
	//Adds a private key to the SSH Agent.
	SSHAgent() (agent.Agent, error)
	//Adds a private key to the SSH Agent.
	SSHAgentAddKey(
		privateKeyPath string,
	) error
	//Removes a private key to the SSH Agent.
	SSHAgentRemoveKey() error
	//Opens an SSH connection to a host.
	SSHConnection(
		host string,
		username string,
	) error
}

func CreateSSHUtils() SSHUtils {
	return &sshUtilsImpl{
		runtime: runtime.CreateRuntime(),
	}
}

type sshUtilsImpl struct{
	runtime runtime.Runtime
}

func (s *sshUtilsImpl) LocalPortForward(
	localPort uint,
	remoteAddr string,
	remotePort uint,
	host string,
	username string,
) error {
	a, err := s.SSHAgent()

	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{
			ssh.PublicKeysCallback(a.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	localListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", remoteAddr, localPort))

	if err != nil {
		return errors.Wrapf(err, "failed to listen on %s:%d", remoteAddr, localPort)
	}

	errChan := make(chan error)

	go func() {
		for err := range errChan {
			log.Println(err)
		}
	}()

	for {
		//For each local connection
		localConn, err := localListener.Accept()

		if err != nil {
			return errors.Wrapf(err, "failed to accept connection on %s", remoteAddr)
		}

		go s.forward(
			localConn,
			remoteAddr,
			remotePort,
			config,
			host,
			errChan,
		)
	}
}

//Returns an SSH Agent instance.
func (s *sshUtilsImpl) SSHAgent() (agent.Agent, error) {
	if s.runtime.CurrentOS() == "windows" {

		conn, err := winio.DialPipe(windowsSSHAgentPipe, nil)

		if err != nil {
			return nil, errors.Wrap(err, "failed to open ssh agent pipe")
		}

		return agent.NewClient(conn), nil
	}

	return nil, errors.New("SSH agent not supported for this OS")
}

//Adds a private key to the SSH Agent.
func (s *sshUtilsImpl) SSHAgentAddKey(
	keyPairPath string,
) error {
	a, err := s.SSHAgent()

	if err != nil {
		return errors.Wrap(err, "failed to get the SSH agent")
	}

	keyPairPEM, err := ioutil.ReadFile(keyPairPath)

	if err != nil {
		return err
	}

	block, _ := pem.Decode(keyPairPEM)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return err
	}

	if err:= a.Add(agent.AddedKey{
		Comment: dockerRemoteEC2KeyID,
		PrivateKey:           key,
	}); err != nil {
		return errors.Wrap(err, "failed to add the key to the agent")
	}

	return nil
}

//Removes a private key to the SSH Agent.
func (s *sshUtilsImpl) SSHAgentRemoveKey() error {
	a, err := s.SSHAgent()

	if err != nil {
		return errors.Wrap(err, "failed to get the SSH agent")
	}

	keys, err := a.List()

	if err != nil {
		return err
	}

	for _, key := range keys {

		if key.Comment == dockerRemoteEC2KeyID {
			publicKey, err := ssh.ParsePublicKey(key.Blob)

			if err != nil {
				return err
			}

			return a.Remove(publicKey)
		}
	}

	return nil
}
//Opens an SSH connection to a host.
func (s *sshUtilsImpl) SSHConnection(
	host string,
	username string,
) error {
	a, err := s.SSHAgent()

	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{
			ssh.PublicKeysCallback(a.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)

	if err != nil {
		return errors.Wrap(err, "failed to dial with the host")
	}

	session, err := conn.NewSession()

	if err != nil {
		return errors.Wrap(err, "failed to create the SSH session")
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
		return errors.Wrap(err, "failed to create SSH the stdin pipe")
	}

	go io.Copy(stdin, os.Stdin)

	stdinFd := int(os.Stdin.Fd())
	stdoutFd := int(os.Stdout.Fd())

	if terminal.IsTerminal(stdinFd) {
		originalState, err := terminal.MakeRaw(stdinFd)

		if err != nil {
			return errors.Wrap(err, "failed to create terminal original state")
		}

		defer terminal.Restore(stdinFd, originalState)

		termWidth, termHeight, err := terminal.GetSize(stdoutFd)

		if err != nil {
			return errors.Wrap(err, "failed to get the terminal size")
		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)

		if err != nil {
			return err
		}
	}

	stdout, err := session.StdoutPipe()

	if err != nil {
		return errors.Wrap(err, "failed to create the stdout pipe")
	}

	go io.Copy(os.Stdout, stdout)

	stderr, err := session.StderrPipe()

	if err != nil {
		return errors.Wrap(err, "failed to create the stderr pipe")
	}

	go io.Copy(os.Stderr, stderr)

	err = session.Shell()

	if err != nil {
		return errors.Wrap(err, "failed to create the shell")
	}

	session.Wait()

	return nil
}

func (s *sshUtilsImpl) forward(
	localConn net.Conn,
	remoteAddr string,
	remotePort uint,
	config *ssh.ClientConfig,
	host string,
	errChan chan error,
) {
	sshClientConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)

	if err != nil {
		errChan <- errors.Wrapf(err, "failed to open ssh connection with %s", host)
		return
	}

	remoteAddr = fmt.Sprintf("%s:%d", remoteAddr, remotePort)

	remoteConn, err := sshClientConn.Dial("tcp", remoteAddr)

	if err != nil {
		errChan <- errors.Wrapf(err, "failed to open connection with %s", remoteAddr)
		return
	}

	go func() {
		if _, err := io.Copy(remoteConn, localConn); err != nil {
			errChan <- errors.Wrap(err, "failed to forward connection (remote -> local)")
		}
	}()

	go func() {
		if _, err := io.Copy(localConn, remoteConn); err != nil {
			errChan <- errors.Wrap(err, "failed to forward connection (local -> remote)")
		}
	}()

}

func (s *sshUtilsImpl) keyPairExtractSSHPublicKey(path string) (ssh.AuthMethod, error) {
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
