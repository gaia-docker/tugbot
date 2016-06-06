package actions

import (
	"github.com/gaia-docker/tugbot/container"
	"github.com/gaia-docker/tugbot/container/mockclient"
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestRun(t *testing.T) {

	cc1 := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true"},
	}
	c1 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c1",
			Config: cc1,
		},
		nil,
	)
	cc2 := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTest: "true", container.LabelRunTimestamp: "2016-06-05 16:48:01"},
	}
	c2 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c2",
			Config: cc2,
		},
		nil,
	)
	c3 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c3",
			Config: &dockerclient.ContainerConfig{},
		},
		nil,
	)
	cc4 := &dockerclient.ContainerConfig{
		Labels: map[string]string{container.LabelTugbot: "true"},
	}
	c4 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Name:   "c4",
			Config: cc4,
		},
		nil,
	)
	cs := []container.Container{}
	containers := []container.Container{c1, c2, c3, c4}
	client := mockclient.NewMockClient()
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).
		Run(func(args mock.Arguments) {
			for _, currContainer := range containers {
				if args.Get(0).(container.Filter)(currContainer) {
					cs = append(cs, currContainer)
				}
			}
		}).
		Return(cs, nil)

	Run(client, []string{})
}
