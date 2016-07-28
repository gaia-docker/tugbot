package container

import (
	"fmt"
	"strings"

	"github.com/samalba/dockerclient"
)

// Docker container labels
const (
	LabelTugbot      = "tugbot.service"
	LabelTest        = "tugbot.test"
	LabelEvents      = "tugbot.event.docker"
	LabelCreatedFrom = "tugbot.created.from"
	LabelStopSignal  = "tugbot.stop-signal"
	LabelZodiac      = "tugbot.zodiac.original-image"
)

// NewContainer returns a new Container instance instantiated with the
// specified ContainerInfo and ImageInfo structs.
func NewContainer(containerInfo *dockerclient.ContainerInfo, imageInfo *dockerclient.ImageInfo) *Container {
	return &Container{
		containerInfo: containerInfo,
		imageInfo:     imageInfo,
	}
}

// Container represents a running Docker container.
type Container struct {
	containerInfo *dockerclient.ContainerInfo
	imageInfo     *dockerclient.ImageInfo
}

// ID returns the Docker container ID.
func (c Container) ID() string {
	return c.containerInfo.Id
}

// Name returns the Docker container name.
func (c Container) Name() string {
	return strings.TrimPrefix(c.containerInfo.Name, "/")
}

// ImageID returns the ID of the Docker image that was used to start the
// container.
func (c Container) ImageID() string {
	return c.imageInfo.Id
}

// ImageName returns the name of the Docker image that was used to start the
// container. If the original image was specified without a particular tag, the
// "latest" tag is assumed.
func (c Container) ImageName() string {
	// Compatibility w/ Zodiac deployments
	imageName, ok := c.containerInfo.Config.Labels[LabelZodiac]
	if !ok {
		imageName = c.containerInfo.Config.Image
	}

	if !strings.Contains(imageName, ":") {
		imageName = fmt.Sprintf("%s:latest", imageName)
	}

	return imageName
}

// IsTugbot returns whether or not the current container is the tugbot container itself.
// The tugbot container is identified by the presence of the "tugbot.service"
// label in the container metadata.
func (c Container) IsTugbot() bool {
	val, ok := c.containerInfo.Config.Labels[LabelTugbot]

	return ok && val == "true"
}

// StopSignal returns the custom stop signal (if any) that is encoded in the
// container's metadata. If the container has not specified a custom stop
// signal, the empty string "" is returned.
func (c Container) StopSignal() string {
	if val, ok := c.containerInfo.Config.Labels[LabelStopSignal]; ok {
		return val
	}

	return ""
}

// IsTugbotCandidate returns whether or not a container is a candidate to run by tugbot.
// A candidate container is identified by the presence of "tugbot.test",
// it doesn't contain "tugbot.created.from" in the container metadata and it state is "Exited".
func (c Container) IsTugbotCandidate() bool {
	ret := false
	val, ok := c.containerInfo.Config.Labels[LabelTest]
	if ok && val == "true" {
		if !c.IsCreatedByTugbot() {
			ret = c.containerInfo.State.StateString() == "exited"
		}
	}

	return ret
}

// IsCreatedByTugbot returns whether or not a container created by tugbot.
func (c Container) IsCreatedByTugbot() bool {
	val, ok := c.containerInfo.Config.Labels[LabelCreatedFrom]

	return ok && val != ""
}

// IsEventListener returns whether or not a container should run when an event e is occurred.
func (c Container) IsEventListener(e *dockerclient.Event) bool {
	ret := false
	if e != nil {
		events, ok := c.containerInfo.Config.Labels[LabelEvents]
		ret = ok && sliceContains(e.Status, strings.Split(events, ","))
	}

	return ret
}

// Any links in the HostConfig need to be re-written before they can be
// re-submitted to the Docker create API.
func (c Container) hostConfig() *dockerclient.HostConfig {
	hostConfig := c.containerInfo.HostConfig

	for i, link := range hostConfig.Links {
		name := link[0:strings.Index(link, ":")]
		alias := link[strings.LastIndex(link, "/"):]

		hostConfig.Links[i] = fmt.Sprintf("%s:%s", name, alias)
	}

	return hostConfig
}
