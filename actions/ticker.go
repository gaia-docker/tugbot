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
					ID:        currTaskId,
					Name:      currCandidate.Name(),
					Job:       startContainerFrom,
					JobParams: []interface{}{client, currCandidate},
					Interval:  interval}
				tasks = append(tasks, currTaskId)
				if ok := manager.RunNewRecurringTask(currTask); ok {
					log.Infof("Ticker starting new recuring task... (Container ID: %s, Name: %s, Interval: %s)", currTaskId, currTask.Name, currTask.Interval)
				}
			}
		}
		manager.Refresh(tasks)
	}
}

func startContainerFrom(params []interface{}) error {
	client := params[0].(container.Client)
	c := params[1].(container.Container)
	if _, err := client.Inspect(c.ID()); err != nil {
		return err
	}

	return client.StartContainerFrom(c)
}
