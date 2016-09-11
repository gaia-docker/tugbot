package container

import (
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

// IsCreatedByTugbot - true if created by tugbot
func IsCreatedByTugbot(e *dockerclient.Event) bool {
	ret := false
	if "" != e.Actor.Attributes[LabelCreatedFrom] {
		ret = true
	}

	return ret
}

// IsSwarmTask - true if container is a swarm task
func IsSwarmTask(e *dockerclient.Event) bool {
	ret := false
	taskID := e.Actor.Attributes[LabelDockerSwarmTaskID]
	if "" != taskID {
		log.Debugf("Swarm service task event (task ID: %s, %v)", taskID, e)
		ret = true
	}

	return ret
}
