package container

import (
	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

func IsCreatedByTugbot(e *dockerclient.Event) bool {
	ret := false
	if "" != e.Actor.Attributes[LabelCreatedFrom] {
		ret = true
	}

	return ret
}

func IsSwarmTask(e *dockerclient.Event) bool {
	ret := false
	taskId := e.Actor.Attributes[LabelDockerSwarmTaskId]
	if "" != taskId {
		log.Debugf("Swarm service task event (task ID: %s, %v)", taskId, e)
		ret = true
	}

	return ret
}
