package container

import (
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"

	"fmt"
	"strings"
	"time"
)

// Docker container labels
const (
	TugbotService     = "tugbot.service"
	TugbotTest        = "tugbot.test"
	TugbotEventDocker = "tugbot.event.docker"
	TugbotEventTimer  = "tugbot.event.timer"
	TugbotCreatedFrom = "tugbot.created.from"
	SwarmTaskID       = "com.docker.swarm.task.id"
)

// Docker Event Filter
const (
	// type filter: tugbot.event.docker.filter.type=container|image|daemon|network|volume|plugin
	TypeFilter = "tugbot.event.docker.filter.type"
	// action filter (depends on type), for 'container' type:
	//  - attach, commit, copy, create, destroy, detach, die, exec_create, exec_detach, exec_start, export,
	//  - health_status, kill, oom, pause, rename, resize, restart, start, stop, top, unpause, update
	ActionFilter = "tugbot.event.docker.filter.action"
	// container filter: use name, comma separated name list or RE2 regexp
	ContainerFilter = "tugbot.event.docker.filter.container"
	// image filter: use name, comma separated name list or RE2 regexp
	ImageFilter = "tugbot.event.docker.filter.image"
	// label filter: use key=value comma separated pairs
	LabelFilter = "tugbot.event.docker.filter.label"
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
	imageName := c.containerInfo.Config.Image

	if !strings.Contains(imageName, ":") {
		imageName = fmt.Sprintf("%s:latest", imageName)
	}

	return imageName
}

// IsTugbot returns whether or not the current container is the tugbot container itself.
// The tugbot container is identified by the presence of the "tugbot.service"
// label in the container metadata.
func (c Container) IsTugbot() bool {
	val, ok := c.containerInfo.Config.Labels[TugbotService]

	return ok && val == "true"
}

// IsTugbotCandidate returns whether or not a container is a candidate to run by tugbot.
// A candidate container is identified by the presence of "tugbot.test",
// it doesn't contain "tugbot.created.from" in the container metadata and it state is "Exited".
func (c Container) IsTugbotCandidate() bool {
	ret := false
	val, ok := c.containerInfo.Config.Labels[TugbotTest]
	if ok && val == "true" {
		if !c.IsCreatedByTugbot() {
			ret = c.containerInfo.State.StateString() == "exited"
		}
	}

	return ret
}

// IsCreatedByTugbot returns whether or not a container created by tugbot.
func (c Container) IsCreatedByTugbot() bool {
	val, ok := c.containerInfo.Config.Labels[TugbotCreatedFrom]

	return ok && val != ""
}

// IsEventListener returns whether or not a container should run when an event e is occurred.
func (c Container) IsEventListener(e *dockerclient.Event) bool {
	ret := false
	if e != nil {
		// check if container is subscribed to Docker events, i.e. 'tugbot.event.docker' label exists
		_, ret = c.containerInfo.Config.Labels[TugbotEventDocker]
		if ret {
			// filter by event type
			if typeFilter, ok := c.containerInfo.Config.Labels[TypeFilter]; ok {
				ret = sliceContains(e.Type, splitAndTrimSpaces(typeFilter, ","))
			}
			// filter by event action
			if actionFilter, ok := c.containerInfo.Config.Labels[ActionFilter]; ok {
				ret = ret && sliceContains(e.Action, splitAndTrimSpaces(actionFilter, ","))
			}
			// filter by container name or name regexp
			if containerFilter, ok := c.containerInfo.Config.Labels[ContainerFilter]; ok {
				ret = ret && inFilterOrList(e.Actor.Attributes["name"], containerFilter)
			}
		}
		// filter by event image
		if imageFilter, ok := c.containerInfo.Config.Labels[ImageFilter]; ok {
			// get image name from event.From field
			imageName := e.From
			// in case of "image" event.Type, event.ID contains image ID (name:tag) for 'pull' action and sha256:num for untag and delete
			if e.Type == "image" {
				imageName = e.ID
			}
			ret = ret && inFilterOrList(imageName, imageFilter)
		}
		// filter by event labels
		if labelFilter, ok := c.containerInfo.Config.Labels[LabelFilter]; ok {
			labels := splitAndTrimSpaces(labelFilter, ",")
			for _, label := range labels {
				ret = ret && mapContains(e.Actor.Attributes, splitAndTrimSpaces(label, "="))
			}
		}
	}

	return ret
}

// GetEventListenerTimer returns interval duration between a test container run and true
// if docker label exist and label value parsed into Duration, Otherwise false.
func (c Container) GetEventListenerInterval() (time.Duration, bool) {
	var ret time.Duration
	val, ok := c.containerInfo.Config.Labels[TugbotEventTimer]
	if ok {
		interval, err := time.ParseDuration(val)
		if err != nil {
			log.Errorf("Failed to parse %s docker label: %s into golang Duration (%v)", TugbotEventTimer, val, err)
			ok = false
		} else {
			ret = interval
		}
	}

	return ret, ok
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
