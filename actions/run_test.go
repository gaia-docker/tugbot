package actions

import (
	"testing"
	"time"
	"errors"
	
	"github.com/gaia-docker/tugbot/container"
	"github.com/gaia-docker/tugbot/container/mockclient"
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var stateExited = &dockerclient.State{Running: false, Dead: false, StartedAt: time.Now()}

func TestRun(t *testing.T) {
	cc1 := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true"},
	}
	c1 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c1",
			Config: cc1,
			State:  stateExited,
		},
		nil,
	)

	client := mockclient.NewMockClient()
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{c1}, nil)
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c1.Name(), args.Get(0).(container.Container).Name())
		}).Return(nil)

	err := Run(client, []string{})
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestRun_NoCandidates(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{}, nil)

	Run(client, []string{})
	client.AssertExpectations(t)
}

func TestRun_ErrorListContainers(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{}, errors.New("whoops"))

	err := Run(client, []string{})
	assert.Error(t, err)
	assert.EqualError(t, err, "whoops")
	client.AssertExpectations(t)
}

func TestRun_ErrorStartContainerFrom(t *testing.T) {
	cc1 := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true"},
	}
	c1 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c1",
			Config: cc1,
			State:  stateExited,
		},
		nil,
	)

	client := mockclient.NewMockClient()
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{c1}, nil)
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c1.Name(), args.Get(0).(container.Container).Name())
		}).Return(errors.New("whoops"))

	err := Run(client, []string{})
	assert.NoError(t, err)
	client.AssertExpectations(t)
}
