package container

import (
	"crypto/tls"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

// A Filter is a prototype for a function that can be used to filter the
// results from a call to the ListContainers() method on the Client.
type Filter func(Container) bool

// A Client is the interface through which tugbot interacts with the
// Docker API.
type Client interface {
	ListContainers(Filter) ([]Container, error)
	StartContainerFrom(Container) error
	StartMonitorEvents(dockerclient.Callback)
	StopAllMonitorEvents()
}

// NewClient returns a new Client instance which can be used to interact with
// the Docker API.
func NewClient(dockerHost string, tlsConfig *tls.Config, pullImages bool) Client {
	docker, err := dockerclient.NewDockerClient(dockerHost, tlsConfig)
	if err != nil {
		log.Fatalf("Error instantiating Docker client: %s", err)
	}

	return dockerClient{api: docker}
}

type dockerClient struct {
	api dockerclient.Client
}

func (client dockerClient) ListContainers(fn Filter) ([]Container, error) {
	cs := []Container{}

	log.Debug("Retrieving containers")

	containers, err := client.api.ListContainers(true, false, "")
	if err != nil {
		return nil, err
	}

	for _, currContainer := range containers {
		c, err := client.toContainer(currContainer.Id)
		if err != nil {
			continue
		}
		if fn(*c) {
			cs = append(cs, *c)
		}
	}

	return cs, nil
}

func (client dockerClient) StartContainerFrom(c Container) error {
	config := c.containerInfo.Config
	hostConfig := c.hostConfig()
	name := c.Name()
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	config.Labels[TugbotCreatedFrom] = name

	log.Debugf("Starting container from %s", name)
	var err error
	var newContainerID string
	newContainerName := fmt.Sprintf("tugbot_%s_%s", name, time.Now().Format("20060102150405"))
	newContainerID, err = client.api.CreateContainer(config, newContainerName, nil)
	if err != nil {
		return err
	}

	log.Infof("Starting container %s (%s)", newContainerName, newContainerID)

	return client.api.StartContainer(newContainerID, hostConfig)
}

func (client dockerClient) StartMonitorEvents(cb dockerclient.Callback) {
	client.api.StartMonitorEvents(cb, nil)
}

func (client dockerClient) StopAllMonitorEvents() {
	client.api.StopAllMonitorEvents()
}

func (client dockerClient) toContainer(containerID string) (*Container, error) {
	containerInfo, err := client.api.InspectContainer(containerID)
	if err != nil {
		log.Errorf("Failed retrieving container info (%s). Error: %+v", containerID, err)
		return nil, err
	}

	imageInfo, err := client.api.InspectImage(containerInfo.Image)
	if err != nil {
		log.Errorf("Failed retrieving image info (%s). Error: %+v", containerInfo.Image, err)
		return nil, err
	}

	return &Container{containerInfo: containerInfo, imageInfo: imageInfo}, nil
}
