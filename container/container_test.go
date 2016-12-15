package container

import (
	"testing"
	"time"

	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
)

var stateExited = &dockerclient.State{Running: false, Dead: false, StartedAt: time.Now()}

func TestID(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{Id: "foo"},
	}

	assert.Equal(t, "foo", c.ID())
}

func TestName(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{Name: "foo"},
	}

	assert.Equal(t, "foo", c.Name())
}

func TestNameStartWithSlash(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{Name: "/foo"},
	}

	assert.Equal(t, "foo", c.Name())
}

func TestImageID(t *testing.T) {
	c := Container{
		imageInfo: &dockerclient.ImageInfo{
			Id: "foo",
		},
	}

	assert.Equal(t, "foo", c.ImageID())
}

func TestImageName_Tagged(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Image: "foo:latest",
			},
		},
	}

	assert.Equal(t, "foo:latest", c.ImageName())
}

func TestImageName_Untagged(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Image: "foo",
			},
		},
	}

	assert.Equal(t, "foo:latest", c.ImageName())
}

func TestIsTugbot_True(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotService: "true"},
			},
		},
	}

	assert.True(t, c.IsTugbot())
}

func TestIsTugbot_WrongLabelValue(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotService: "false"},
			},
		},
	}

	assert.False(t, c.IsTugbot())
}

func TestIsTugbot_NoLabel(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{},
			},
		},
	}

	assert.False(t, c.IsTugbot())
}

func TestIsTugbotCandidate_True(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotTest: "true"},
			},
		},
	}

	assert.True(t, c.IsTugbotCandidate())
}

func TestIsTugbotCandidate_TrueRunTimestampLabelEmpty(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotTest: "true", TugbotCreatedFrom: ""},
			},
		},
	}

	assert.True(t, c.IsTugbotCandidate())
}

func TestIsTugbotCandidate_FalseRunningState(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: &dockerclient.State{Running: true, Dead: false, StartedAt: time.Now()},
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotTest: "true"},
			},
		},
	}

	assert.False(t, c.IsTugbotCandidate())
}

func TestIsTugbotCandidate_FalseNoLabels(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State:  stateExited,
			Config: &dockerclient.ContainerConfig{},
		},
	}

	assert.False(t, c.IsTugbotCandidate())
}

func TestIsTugbotCandidate_FalseIncludeRunTimestampLabel(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotTest: "true", TugbotCreatedFrom: "2016-06-05 16:48:01.9042582 +0300 IDT"},
			},
		},
	}

	assert.False(t, c.IsTugbotCandidate())
}

func TestIsEventListener_True(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{
					TugbotTest:        "true",
					TugbotEventDocker: "",
					TypeFilter:        "container",
					ActionFilter:      "start",
				},
			},
		},
	}

	assert.True(t, c.IsEventListener(&dockerclient.Event{Type: "container", Action: "start"}))
}

func TestIsEventListener_False(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotTest: "true", TugbotEventDocker: "", TypeFilter: "container", ActionFilter: "create,start,destroy"},
			},
		},
	}

	assert.False(t, c.IsEventListener(&dockerclient.Event{Action: "unexpected"}))
}

func TestContainerIsCreatedByTugbot_Ture(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotTest: "true", TugbotEventDocker: "", TypeFilter: "container", ActionFilter: "create,start,destroy", TugbotCreatedFrom: "aabb"},
			},
		},
	}

	assert.True(t, c.IsCreatedByTugbot())
}

func TestContainerIsCreatedByTugbot_False(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State:  stateExited,
			Config: &dockerclient.ContainerConfig{},
		},
	}

	assert.False(t, c.IsCreatedByTugbot())
}

func TestGetEventListenerTimer(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotEventTimer: "7s"},
			},
		},
	}
	interval, ok := c.GetEventListenerInterval()

	assert.True(t, ok)
	assert.Equal(t, interval, time.Second*7)
}

func TestGetEventListenerTimer_LabelNotFound(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{},
		},
	}
	_, ok := c.GetEventListenerInterval()

	assert.False(t, ok)
}

func TestGetEventListenerTimer_FailedToParseInterval(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{TugbotEventTimer: "7"},
			},
		},
	}
	_, ok := c.GetEventListenerInterval()

	assert.False(t, ok)
}
