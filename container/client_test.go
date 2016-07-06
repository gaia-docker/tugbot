package container

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/samalba/dockerclient"
	"github.com/samalba/dockerclient/mockclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func allContainers(Container) bool {
	return true
}

func noContainers(Container) bool {
	return false
}

func TestListContainers_Success(t *testing.T) {
	ci := &dockerclient.ContainerInfo{Image: "abc123", Config: &dockerclient.ContainerConfig{Image: "img"}}
	ii := &dockerclient.ImageInfo{}
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{{Id: "foo", Names: []string{"bar"}}}, nil)
	api.On("InspectContainer", "foo").Return(ci, nil)
	api.On("InspectImage", "abc123").Return(ii, nil)

	client := dockerClient{api: api}
	cs, err := client.ListContainers(allContainers)

	assert.NoError(t, err)
	assert.Len(t, cs, 1)
	assert.Equal(t, ci, cs[0].containerInfo)
	assert.Equal(t, ii, cs[0].imageInfo)
	api.AssertExpectations(t)
}

func TestListContainers_Filter(t *testing.T) {
	ci := &dockerclient.ContainerInfo{Image: "abc123", Config: &dockerclient.ContainerConfig{Image: "img"}}
	ii := &dockerclient.ImageInfo{}
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{{Id: "foo", Names: []string{"bar"}}}, nil)
	api.On("InspectContainer", "foo").Return(ci, nil)
	api.On("InspectImage", "abc123").Return(ii, nil)

	client := dockerClient{api: api}
	cs, err := client.ListContainers(noContainers)

	assert.NoError(t, err)
	assert.Len(t, cs, 0)
	api.AssertExpectations(t)
}

func TestListContainers_ListError(t *testing.T) {
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{}, errors.New("oops"))

	client := dockerClient{api: api}
	_, err := client.ListContainers(allContainers)

	assert.Error(t, err)
	assert.EqualError(t, err, "oops")
	api.AssertExpectations(t)
}

func TestListContainers_InspectContainerError(t *testing.T) {
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{{Id: "foo", Names: []string{"bar"}}}, nil)
	api.On("InspectContainer", "foo").Return(&dockerclient.ContainerInfo{}, errors.New("uh-oh"))

	client := dockerClient{api: api}
	cs, err := client.ListContainers(allContainers)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(cs))
	api.AssertExpectations(t)
}

func TestListContainers_InspectImageError(t *testing.T) {
	ci := &dockerclient.ContainerInfo{Image: "abc123", Config: &dockerclient.ContainerConfig{Image: "img"}}
	ii := &dockerclient.ImageInfo{}
	api := mockclient.NewMockClient()
	api.On("ListContainers", true, false, "").Return([]dockerclient.Container{{Id: "foo", Names: []string{"bar"}}}, nil)
	api.On("InspectContainer", "foo").Return(ci, nil)
	api.On("InspectImage", "abc123").Return(ii, errors.New("whoops"))

	client := dockerClient{api: api}
	cs, err := client.ListContainers(allContainers)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(cs))
	api.AssertExpectations(t)
}

func TestStopContainer_DefaultSuccess(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:   "foo",
			Id:     "abc123",
			Config: &dockerclient.ContainerConfig{},
		},
	}

	ci := &dockerclient.ContainerInfo{
		State: &dockerclient.State{
			Running: false,
		},
	}

	api := mockclient.NewMockClient()
	api.On("KillContainer", "abc123", "SIGTERM").Return(nil)
	api.On("InspectContainer", "abc123").Return(ci, nil).Once()
	api.On("RemoveContainer", "abc123", true, false).Return(nil)
	api.On("InspectContainer", "abc123").Return(&dockerclient.ContainerInfo{}, errors.New("Not Found"))

	client := dockerClient{api: api}
	err := client.StopContainer(c, time.Second)

	assert.NoError(t, err)
	api.AssertExpectations(t)
}

func TestStopContainer_CustomSignalSuccess(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name: "foo",
			Id:   "abc123",
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{LabelStopSignal: "SIGUSR1"}},
		},
	}

	ci := &dockerclient.ContainerInfo{
		State: &dockerclient.State{
			Running: false,
		},
	}

	api := mockclient.NewMockClient()
	api.On("KillContainer", "abc123", "SIGUSR1").Return(nil)
	api.On("InspectContainer", "abc123").Return(ci, nil).Once()
	api.On("RemoveContainer", "abc123", true, false).Return(nil)
	api.On("InspectContainer", "abc123").Return(&dockerclient.ContainerInfo{}, errors.New("Not Found"))

	client := dockerClient{api: api}
	err := client.StopContainer(c, time.Second)

	assert.NoError(t, err)
	api.AssertExpectations(t)
}

func TestStopContainer_KillContainerError(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:   "foo",
			Id:     "abc123",
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	api.On("KillContainer", "abc123", "SIGTERM").Return(errors.New("oops"))

	client := dockerClient{api: api}
	err := client.StopContainer(c, time.Second)

	assert.Error(t, err)
	assert.EqualError(t, err, "oops")
	api.AssertExpectations(t)
}

func TestStopContainer_RemoveContainerError(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:   "foo",
			Id:     "abc123",
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	api.On("KillContainer", "abc123", "SIGTERM").Return(nil)
	api.On("InspectContainer", "abc123").Return(&dockerclient.ContainerInfo{}, errors.New("dangit"))
	api.On("RemoveContainer", "abc123", true, false).Return(errors.New("whoops"))

	client := dockerClient{api: api}
	err := client.StopContainer(c, time.Second)

	assert.Error(t, err)
	assert.EqualError(t, err, "whoops")
	api.AssertExpectations(t)
}

func TestStartContainerFrom_Success(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:       "foo",
			Config:     &dockerclient.ContainerConfig{},
			HostConfig: &dockerclient.HostConfig{},
		},
		imageInfo: &dockerclient.ImageInfo{
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	api.On("CreateContainer",
		mock.MatchedBy(func(config *dockerclient.ContainerConfig) bool {
			return config.Labels[LabelCreatedFrom] == "foo"
		}),
		mock.MatchedBy(func(name string) bool {
			return strings.HasPrefix(name, "tugbot_foo_")
		}),
		mock.AnythingOfType("*dockerclient.AuthConfig")).Return("def789", nil).Once()
	api.On("StartContainer", "def789", mock.AnythingOfType("*dockerclient.HostConfig")).Return(nil).Once()

	client := dockerClient{api: api}
	err := client.StartContainerFrom(c)

	assert.NoError(t, err)
	api.AssertExpectations(t)
}

func TestStartContainerFrom_SuccessUsingAuthConfig(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:       "foo",
			Config:     &dockerclient.ContainerConfig{},
			HostConfig: &dockerclient.HostConfig{},
		},
		imageInfo: &dockerclient.ImageInfo{
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	originalUsername := username
	originalPassword := password
	originalEmail := email
	username = "user-test"
	password = "123456"
	email = "user-test@hpe.com"
	defer func() {
		username = originalUsername
		password = originalPassword
		email = originalEmail
	}()
	api.On("CreateContainer",
		mock.MatchedBy(func(config *dockerclient.ContainerConfig) bool {
			return config.Labels[LabelCreatedFrom] == "foo"
		}),
		mock.MatchedBy(func(name string) bool {
			return strings.HasPrefix(name, "tugbot_foo_")
		}),
		mock.MatchedBy(func(authConfig *dockerclient.AuthConfig) bool {
			return "user-test" == authConfig.Username && "123456" == authConfig.Password && "user-test@hpe.com" == authConfig.Email
		})).Return("def789", nil).Once()
	api.On("StartContainer", "def789", mock.AnythingOfType("*dockerclient.HostConfig")).Return(nil).Once()

	client := dockerClient{api: api}
	err := client.StartContainerFrom(c)

	assert.NoError(t, err)
	api.AssertExpectations(t)
}

func TestStartContainerFrom_CreateContainerError(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:       "foo",
			Config:     &dockerclient.ContainerConfig{},
			HostConfig: &dockerclient.HostConfig{},
		},
		imageInfo: &dockerclient.ImageInfo{
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	api.On("CreateContainer",
		mock.MatchedBy(func(config *dockerclient.ContainerConfig) bool {
			return config.Labels[LabelCreatedFrom] == "foo"
		}),
		mock.MatchedBy(func(name string) bool {
			return strings.HasPrefix(name, "tugbot_foo_")
		}), mock.AnythingOfType("*dockerclient.AuthConfig")).Return("", errors.New("oops")).Once()

	client := dockerClient{api: api}
	err := client.StartContainerFrom(c)

	assert.Error(t, err)
	assert.EqualError(t, err, "oops")
	api.AssertExpectations(t)
}

func TestStartContainerFrom_StartContainerError(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Name:       "foo",
			Config:     &dockerclient.ContainerConfig{},
			HostConfig: &dockerclient.HostConfig{},
		},
		imageInfo: &dockerclient.ImageInfo{
			Config: &dockerclient.ContainerConfig{},
		},
	}

	api := mockclient.NewMockClient()
	api.On("CreateContainer",
		mock.MatchedBy(func(config *dockerclient.ContainerConfig) bool {
			return config.Labels[LabelCreatedFrom] == "foo"
		}),
		mock.MatchedBy(func(name string) bool {
			return strings.HasPrefix(name, "tugbot_foo_")
		}),
		mock.AnythingOfType("*dockerclient.AuthConfig")).Return("created-container-id", nil).Once()
	api.On("StartContainer", "created-container-id", mock.Anything).Return(errors.New("whoops")).Once()

	client := dockerClient{api: api}
	err := client.StartContainerFrom(c)

	assert.Error(t, err)
	assert.EqualError(t, err, "whoops")
	api.AssertExpectations(t)
}

func TestClientIsCreatedByTugbot_True(t *testing.T) {
	attributes := map[string]string{LabelTest: "true", LabelCreatedFrom: "aabb"}
	api := mockclient.NewMockClient()
	client := dockerClient{api: api}
	created := client.IsCreatedByTugbot(
		&dockerclient.Event{Actor: dockerclient.Actor{Attributes: attributes}})

	assert.True(t, created)
	api.AssertExpectations(t)
}

func TestClientIsCreatedByTugbot_False(t *testing.T) {
	attributes := map[string]string{LabelTest: "true"}
	api := mockclient.NewMockClient()
	client := dockerClient{api: api}
	created := client.IsCreatedByTugbot(
		&dockerclient.Event{Actor: dockerclient.Actor{Attributes: attributes}})

	assert.False(t, created)
	api.AssertExpectations(t)
}

func TestClientIsCreatedByTugbot_NoAttributes(t *testing.T) {
	api := mockclient.NewMockClient()
	client := dockerClient{api: api}
	created := client.IsCreatedByTugbot(&dockerclient.Event{})

	assert.False(t, created)
	api.AssertExpectations(t)
}
