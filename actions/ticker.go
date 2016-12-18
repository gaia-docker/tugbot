package actions

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gaia-docker/tugbot-common"
	"github.com/gaia-docker/tugbot/container"
	"golang.org/x/net/context"

	"time"
)

// RunTickerTestContainers on a clock intervals runs test containers that should run recurring.
func RunTickerTestContainers(ctx context.Context, client container.Client, interval time.Duration) {
	manager := common.NewTaskManager()
	ticker := time.NewTicker(interval)
	for {
		runNewTasks(manager, client)
		select {
		case <-ctx.Done():
			ticker.Stop()
			manager.StopAllTasks()
			log.Info("Test Containers' Ticker Stopped.")

			return
		case <-ticker.C:
			break
		}
	}
}

func runNewTasks(manager common.TaskManager, client container.Client) {
	candidates, err := client.ListContainers(func(c container.Container) bool {
		return c.IsTugbotCandidate()
	})
	if err != nil {
		log.Errorf("Failed to get list test containers candidates for timer event (%v)", err)
	} else {
		var tasks []string
		for _, currCandidate := range candidates {
			interval, ok := currCandidate.GetEventListenerInterval()
			if ok {
				currTaskId := currCandidate.ID()
				currTask := common.Task{
					ID:   currTaskId,
					Name: currCandidate.Name(),
					Job: func(param interface{}) error {
						return client.StartContainerFrom(param.(container.Container))
					},
					JobParam: currCandidate,
					Interval: interval}
				if ok := manager.RunNewRecurringTask(currTask); ok {
					tasks = append(tasks, currTaskId)
				}
			}
		}
		manager.Refresh(tasks)
	}
}
