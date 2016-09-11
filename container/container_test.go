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
				Labels: map[string]string{LabelTugbot: "true"},
			},
		},
	}

	assert.True(t, c.IsTugbot())
}

func TestIsTugbot_WrongLabelValue(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{LabelTugbot: "false"},
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

func TestStopSignal_Present(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{
					LabelStopSignal: "SIGQUIT",
				},
			},
		},
	}

	assert.Equal(t, "SIGQUIT", c.StopSignal())
}

func TestStopSignal_NoLabel(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{},
			},
		},
	}

	assert.Equal(t, "", c.StopSignal())
}

func TestIsTugbotCandidate_True(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{LabelTest: "true"},
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
				Labels: map[string]string{LabelTest: "true", LabelCreatedFrom: ""},
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
				Labels: map[string]string{LabelTest: "true"},
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
				Labels: map[string]string{LabelTest: "true", LabelCreatedFrom: "2016-06-05 16:48:01.9042582 +0300 IDT"},
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
				Labels: map[string]string{LabelTest: "true", LabelEvents: "create,start,destroy"},
			},
		},
	}

	assert.True(t, c.IsEventListener(&dockerclient.Event{Status: "start"}))
}

func TestIsEventListener_False(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{LabelTest: "true", LabelEvents: "create,start,destroy"},
			},
		},
	}

	assert.False(t, c.IsEventListener(&dockerclient.Event{Status: "foo"}))
}

func TestContainerIsCreatedByTugbot_Ture(t *testing.T) {
	c := Container{
		containerInfo: &dockerclient.ContainerInfo{
			State: stateExited,
			Config: &dockerclient.ContainerConfig{
				Labels: map[string]string{LabelTest: "true", LabelEvents: "create,start,destroy", LabelCreatedFrom: "aabb"},
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
