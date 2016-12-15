package actions

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gaia-docker/tugbot-common"
	"github.com/gaia-docker/tugbot/container"
	"golang.org/x/net/context"

	"time"
)

// RunTickerTestContainers on a clock intervals runs test containers that should run recurring.
func RunTickerTestContainers(ctx context.Context, client container.Client) {
	manager := common.NewTaskManager()
	ticker := time.NewTicker(time.Second * 18)
	for {
		candidates, err := client.ListContainers(func(c container.Container) bool {
			return c.IsTugbotCandidate()
		})
		if err != nil {
			log.Errorf("Failed to get list test containers candidates for timer event (%v)", err)
		} else {
			for _, currCandidate := range candidates {
				interval, ok := currCandidate.GetEventListenerInterval()
				if ok {
					manager.RunNewRecurringTask(common.Task{
						ID:   currCandidate.ID(),
						Name: currCandidate.Name(),
						Job: func(param interface{}) error {
							return client.StartContainerFrom(param.(container.Container))
						},
						JobParam: currCandidate,
						Interval: interval})
				}
			}
		}
		select {
		case <-ctx.Done():
			ticker.Stop()
			manager.StopTasks()
			log.Info("Test Containers' Ticker Stopped.")

			return
		case <-ticker.C:
			break
		}
	}
}
