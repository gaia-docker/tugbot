package actions

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/gaia-docker/tugbot/container"
	"github.com/samalba/dockerclient"
)

// Run looks at Docker containers to see if any of the images
// used to start those containers is a test container.
// For each test container it'll create and start a new container according
// to tugbots' labels.
func Run(client container.Client, names []string, e *dockerclient.Event) error {
	log.Debugf("Docker event: %+v", e)
	var cerr error
	if !container.IsSwarmTask(e) && !container.IsCreatedByTugbot(e) {
		candidates, err := client.ListContainers(containerFilter(names))
		if err != nil {
			return err
		}
		for _, candidate := range candidates {
			if candidate.IsEventListener(e) {
				if err := client.StartContainerFrom(candidate); err != nil {
					log.Error(err)
					// combine errors
					if cerr == nil {
						cerr = err
					} else {
						cerr = errors.New(cerr.Error() + err.Error())
					}
				}
			}
		}
	}
	return cerr
}

func containerFilter(names []string) container.Filter {

	return func(c container.Container) bool {
		return nameFilter(names)(c) && c.IsTugbotCandidate()
	}
}

func nameFilter(names []string) container.Filter {

	if len(names) == 0 {
		// all containers
		return func(container.Container) bool {
			return true
		}
	}

	return func(c container.Container) bool {
		for _, name := range names {
			if (name == c.Name()) || (name == c.Name()[1:]) {
				return true
			}
		}
		return false
	}
}
