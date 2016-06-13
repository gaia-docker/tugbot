package container

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

const (
	defaultStopSignal = "SIGTERM"
)

var username = os.Getenv("REPO_USER")
var password = os.Getenv("REPO_PASS")
var email = os.Getenv("REPO_EMAIL")

// A Filter is a prototype for a function that can be used to filter the
// results from a call to the ListContainers() method on the Client.
type Filter func(Container) bool

// A Client is the interface through which tugbot interacts with the
// Docker API.
type Client interface {
	ListContainers(Filter) ([]Container, error)
	StopContainer(Container, time.Duration) error
	StartContainerFrom(Container) error
	StartMonitorEvents(dockerclient.Callback)
	StopAllMonitorEvents()
	IsCreatedByTugbot(string) (bool, error)
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

func (client dockerClient) StopContainer(c Container, timeout time.Duration) error {
	signal := c.StopSignal()
	if signal == "" {
		signal = defaultStopSignal
	}

	log.Infof("Stopping %s (%s) with %s", c.Name(), c.ID(), signal)

	if err := client.api.KillContainer(c.ID(), signal); err != nil {
		return err
	}

	// Wait for container to exit, but proceed anyway after the timeout elapses
	client.waitForStop(c, timeout)

	log.Debugf("Removing container %s", c.ID())

	if err := client.api.RemoveContainer(c.ID(), true, false); err != nil {
		return err
	}

	// Wait for container to be removed. In this case an error is a good thing
	if err := client.waitForStop(c, timeout); err == nil {
		return fmt.Errorf("Container %s (%s) could not be removed", c.Name(), c.ID())
	}

	return nil
}

func (client dockerClient) StartContainerFrom(c Container) error {
	config := c.runtimeConfig()
	hostConfig := c.hostConfig()
	name := c.Name()
	config.Labels[LabelCreatedFrom] = name

	log.Debugf("Starting container from %s", name)
	var err error
	var newContainerID string
	newContainerName := fmt.Sprintf("tugbot_%s_%s", name, time.Now().Format("20060102150405"))
	if username != "" && password != "" && email != "" {
		auth := dockerclient.AuthConfig{
			Username: username,
			Password: password,
			Email:    email,
		}
		newContainerID, err = client.api.CreateContainer(config, newContainerName, &auth)
	} else {
		newContainerID, err = client.api.CreateContainer(config, newContainerName, nil)
	}

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

func (client dockerClient) IsCreatedByTugbot(containerId string) (bool, error) {
	c, err := client.toContainer(containerId)
	if err != nil {
		return true, err
	}

	return c.IsCreatedByTugbot(), nil
}

func (client dockerClient) toContainer(containerId string) (*Container, error) {
	containerInfo, err := client.api.InspectContainer(containerId)
	if err != nil {
		log.Errorf("Failed retriving container info (%s). Error: %+v", containerId, err)
		return nil, err
	}

	imageInfo, err := client.api.InspectImage(containerInfo.Image)
	if err != nil {
		log.Errorf("Failed retriving image info (%s). Error: %+v", containerInfo.Image, err)
		return nil, err
	}

	return &Container{containerInfo: containerInfo, imageInfo: imageInfo}, nil
}

func (client dockerClient) waitForStop(c Container, waitTime time.Duration) error {
	timeout := time.After(waitTime)

	for {
		select {
		case <-timeout:
			return nil
		default:
			if ci, err := client.api.InspectContainer(c.ID()); err != nil {
				return err
			} else if !ci.State.Running {
				return nil
			}
		}

		time.Sleep(1 * time.Second)
	}
}
