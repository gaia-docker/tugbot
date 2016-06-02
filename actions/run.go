package actions

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gaia-docker/tugbot/container"
)

// Run looks at the running Docker containers to see if any of the images
// used to start those containers is a test container.
// For each test container it'll create and start a new container according
// to tugbots' labels
func Run(client container.Client, names []string) error {

	log.Info("Checking containers for running tests")

	containers, err := client.ListContainers(containerFilter(names))
	if err != nil {
		return err
	}

	for _, c := range containers {
		log.Debug(c)
	}

	return nil
}

func allContainersFilter(container.Container) bool { return true }

func containerFilter(names []string) container.Filter {
	if len(names) == 0 {
		return allContainersFilter
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
