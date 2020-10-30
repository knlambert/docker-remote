package host

type DockerHostSystem interface {
	Up() error
}