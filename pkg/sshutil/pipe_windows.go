package sshutil

func DialPipe() (net.Conn, error) {
	return winio.DialPipe(`\\.\pipe\openssh-ssh-agent`, nil)
}