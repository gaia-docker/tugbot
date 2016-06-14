package actions

import (
	"errors"
	"testing"
	"time"

	"github.com/gaia-docker/tugbot/container"
	"github.com/gaia-docker/tugbot/container/mockclient"
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var stateExited = &dockerclient.State{Running: false, Dead: false, StartedAt: time.Now()}

func TestRun_StartEvent(t *testing.T) {
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true", container.LabelEvents: "start"},
	}
	c := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)

	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(false, nil)
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(false, nil)
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{c}, nil)
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c.Name(), args.Get(0).(container.Container).Name())
		}).Return(nil)

	err := Run(client, []string{}, &dockerclient.Event{Status: "start"})
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestRun_EventCreatedByTugbot(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(true, nil)

	err := Run(client, []string{}, &dockerclient.Event{Status: "start"})
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestRun_ImageEvent(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(true, errors.New("container not found"))

	err := Run(client, []string{}, &dockerclient.Event{Status: "start"})
	assert.Error(t, err)
	assert.EqualError(t, err, "container not found")
	client.AssertExpectations(t)
}

func TestRun_NoCandidates(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(false, nil)
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{}, nil)

	Run(client, []string{}, nil)
	client.AssertExpectations(t)
}

func TestRun_ErrorListContainers(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(false, nil)
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{}, errors.New("whoops"))

	err := Run(client, []string{}, nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "whoops")
	client.AssertExpectations(t)
}

func TestRun_ErrorStartContainerFrom(t *testing.T) {
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true", container.LabelEvents: "start"},
	}
	c := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)

	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(false, nil)
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{c}, nil)
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c.Name(), args.Get(0).(container.Container).Name())
		}).Return(errors.New("whoops"))

	err := Run(client, []string{}, &dockerclient.Event{Status: "start"})
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestFilterName_True(t *testing.T) {
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true"},
	}
	c := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)

	assert.True(t, containerFilter([]string{"c1", "c", "c2"})(c))
}

func TestFilterNoName_True(t *testing.T) {
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true"},
	}
	c := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)

	assert.True(t, containerFilter([]string{})(c))
}

func TestFilterName_False(t *testing.T) {
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true"},
	}
	c := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)

	assert.False(t, containerFilter([]string{"blabla"})(c))
}

func TestFilterContainerStateRunning_False(t *testing.T) {
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true"},
	}
	c := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c",
			Config: cc,
			State:  &dockerclient.State{Running: true, Dead: false, StartedAt: time.Now()},
		},
		nil,
	)

	assert.False(t, containerFilter([]string{})(c))
}
