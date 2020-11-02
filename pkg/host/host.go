package host

type DockerHostSystem interface {
	Down() error
	Up() error
	Shell(publicKeyPath *string) error
}
