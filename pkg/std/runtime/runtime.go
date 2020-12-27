package runtime

import "runtime"

type Runtime interface {
	CurrentOS() string
}

func CreateRuntime() Runtime {
	return &runtimeImpl{}
}

type runtimeImpl struct {}

func (r *runtimeImpl) CurrentOS() string {
	return runtime.GOOS
}
