package actions

import (
	"errors"
	"testing"
	"time"

	"fmt"
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
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{c}, nil)
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c.Name(), args.Get(0).(container.Container).Name())
		}).Return(nil)

	err := Run(client, []string{}, &dockerclient.Event{Status: "start", ID: "123"})
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestRun_CallByEventType(t *testing.T) {
	cc1 := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true", container.LabelEvents: "create"},
	}
	c1 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c1",
			Config: cc1,
			State:  stateExited,
		},
		nil,
	)
	cc2 := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true", container.LabelEvents: "start"},
	}
	c2 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c2",
			Config: cc2,
			State:  stateExited,
		},
		nil,
	)

	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(false, nil)
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{c1, c2}, nil)
	var called []string
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			called = append(called, args.Get(0).(container.Container).Name())
		}).Return(nil)

	err := Run(client, []string{c2.Name()}, &dockerclient.Event{Status: "start", ID: "123"})
	fmt.Println("efffi ", called)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(called))
	assert.Equal(t, c2.Name(), called[0])
	client.AssertExpectations(t)
}

func TestRun_EventCreatedByTugbot(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(true, nil)

	err := Run(client, []string{}, &dockerclient.Event{Status: "start", ID: "123"})
	assert.NoError(t, err)
	client.AssertExpectations(t)
}

func TestRun_ImageEvent(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(true, errors.New("container not found"))

	err := Run(client, []string{}, &dockerclient.Event{Status: "start", ID: "123"})
	assert.Error(t, err)
	assert.EqualError(t, err, "container not found")
	client.AssertExpectations(t)
}

func TestRun_NoCandidates(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(false, nil)
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{}, nil)

	Run(client, []string{}, &dockerclient.Event{Status: "start", ID: "123"})
	client.AssertExpectations(t)
}

func TestRun_ErrorListContainers(t *testing.T) {
	client := mockclient.NewMockClient()
	client.On("IsCreatedByTugbot", mock.AnythingOfType("*dockerclient.Event")).Return(false, nil)
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{}, errors.New("whoops"))

	err := Run(client, []string{}, &dockerclient.Event{Status: "start", ID: "123"})
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

	err := Run(client, []string{}, &dockerclient.Event{Status: "start", ID: "123"})
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
