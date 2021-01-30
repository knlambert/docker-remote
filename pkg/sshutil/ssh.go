package sshutil

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type SSHUtils interface {
	LocalPortForward(
		localPort uint,
		remoteAddr string,
		remotePort uint,
		host string,
		username string,
		publicKeyPath string,
	) error
	//Opens an SSH connection to a host.
	SSHConnection(
		host string,
		username string,
		publicKeyPath string,
	) error
}

func CreateSSHUtils() SSHUtils {
	return &sshUtilsImpl{}
}

type sshUtilsImpl struct{}

func (s *sshUtilsImpl) LocalPortForward(
	localPort uint,
	remoteAddr string,
	remotePort uint,
	host string,
	username string,
	publicKeyPath string,
) error {
	publicKey, err := s.publicKey(publicKeyPath)

	if err != nil {
		return errors.Wrapf(err, "failed to load public key %s", publicKeyPath)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{publicKey},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	localListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", remoteAddr, localPort))

	if err != nil {
		return errors.Wrapf(err, "failed to listen on %s:%d", remoteAddr, localPort)
	}

	errChan := make(chan error)

	go func(){
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

//Opens an SSH connection to a host.
func (s *sshUtilsImpl) SSHConnection(
	host string,
	username string,
	publicKeyPath string,
) error {
	key, err := s.publicKey(publicKeyPath)

	if err != nil {
		return errors.Wrapf(err, "failed to load public key %s", publicKeyPath)
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{key},
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
		if _, err := io.Copy(remoteConn, localConn);  err != nil {
			errChan <- errors.Wrap(err, "failed to forward connection (remote -> local)")
		}
	}()

	go func() {
		if _, err := io.Copy(localConn, remoteConn);  err != nil {
			errChan <- errors.Wrap(err, "failed to forward connection (local -> remote)")
		}
	}()

}

func (s *sshUtilsImpl) publicKey (path string) (ssh.AuthMethod, error) {
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