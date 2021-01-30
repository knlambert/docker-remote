package docker

import (
	"github.com/golang/mock/gomock"
	"github.com/knlambert/docker-remote.git/pkg"
	mock_ioutil "github.com/knlambert/docker-remote.git/pkg/mock/std/ioutil"
	mock_os "github.com/knlambert/docker-remote.git/pkg/mock/std/os"
	mock_runtime "github.com/knlambert/docker-remote.git/pkg/mock/std/runtime"
	mock_user "github.com/knlambert/docker-remote.git/pkg/mock/std/user"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func stubbedDocker(ctrl *gomock.Controller) (
	*dockerImpl,
	*mock_ioutil.MockIOUtil,
	*mock_os.MockOS,
	*mock_runtime.MockRuntime,
	*mock_user.MockUser,
) {
	ioMock := mock_ioutil.NewMockIOUtil(ctrl)
	osMock := mock_os.NewMockOS(ctrl)
	runtimeMock := mock_runtime.NewMockRuntime(ctrl)
	userMock := mock_user.NewMockUser(ctrl)

	return &dockerImpl{
		io: ioMock,
		os: osMock,
		runtime: runtimeMock,
		user:    userMock,
	}, ioMock, osMock, runtimeMock, userMock
}

const (
	expectedHomePath = "/home/barney"
)

func TestContextSetFromNothing(t *testing.T) {
	// Tear up.
	ctrl := gomock.NewController(t)
	s, ioMock, osMock, runtimeMock, userMock := stubbedDocker(ctrl)

	expectedDockerHost := "ssh://ec2-admin@127.0.0.1"
	expectedContextName := "remote-1"
	expectedDockerConfigPath := filepath.Join(expectedHomePath, ".docker")
	expectedContextPrint := "9de980aa335df2af27ea4c640c09878ca7d9dd915b7c208974fecf99e63a1403"
	expectedContextFolderPath := filepath.Join(
		expectedDockerConfigPath, "contexts", "meta", expectedContextPrint,
	)
	expectedContextMetadataFilePath := filepath.Join(expectedContextFolderPath, "meta.json")

	userMock.EXPECT().Current().Return(&user.User{
		HomeDir: expectedHomePath,
	}, nil)

	runtimeMock.EXPECT().CurrentOS().Return("linux")

	context := s.createContextMeta(expectedContextName, expectedDockerHost)
	serialized, _ := context.JSON()

	osMock.EXPECT().PathExists(expectedContextFolderPath).Return(false, nil)
	osMock.EXPECT().MkdirAll(expectedContextFolderPath, os.FileMode(0700)).Return(nil)
	ioMock.EXPECT().WriteFile(expectedContextMetadataFilePath, serialized, os.FileMode(0644))

	//Assertions
	err := s.ContextSet("remote-1", expectedDockerHost)
	assert.Nil(t, err)

	ctrl.Finish()
}

func TestContextSetFailWithUnknownOS(t *testing.T) {
	// Tear up.
	ctrl := gomock.NewController(t)
	s, _, _, runtimeMock , userMock := stubbedDocker(ctrl)

	userMock.EXPECT().Current().Return(&user.User{
		HomeDir: expectedHomePath,
	}, nil)

	runtimeMock.EXPECT().CurrentOS().Return("unknown")

	//Assertions
	err := s.ContextSet("remote-1", "127.0.0.1")

	assert.Errorf(t, err, "ContextSet should return an error on unknown")
	assert.Equal(
		t, errors.Cause(err).(*pkg.InternalError).ErrorType, pkg.NotImplemented, "ContextSet error should be of type not implemented",
	)

	ctrl.Finish()
}
