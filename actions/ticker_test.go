package actions

import (
	"github.com/gaia-docker/tugbot/container"
	"github.com/gaia-docker/tugbot/container/mockclient"
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/1.4/_vendor/golang.org/x/net/context"

	"errors"
	"sync"
	"testing"
)

func TestRunTickerTestContainers_FailedToGetListContainers(t *testing.T) {
	touch := false
	var locker sync.Mutex
	var wg1, wg2 sync.WaitGroup
	wg1.Add(1)
	wg2.Add(1)
	client := mockclient.NewMockClient()
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).
		Run(func(args mock.Arguments) {
			locker.Lock()
			if !touch {
				touch = true
				wg1.Done()
			}
			locker.Unlock()
		}).Return([]container.Container{}, errors.New("Expected :)"))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		RunTickerTestContainers(ctx, client)
		wg2.Done()
	}()
	wg1.Wait()
	cancel()
	wg2.Wait()

	assert.True(t, touch)
	client.AssertExpectations(t)

}

func TestRunTickerTestContainers(t *testing.T) {
	touch := false
	var locker sync.Mutex
	var wg1, wg2 sync.WaitGroup
	wg1.Add(1)
	wg2.Add(1)
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			container.TugbotTest:       "true",
			container.TugbotEventTimer: "10s",
		},
	}
	c := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Id:     "02131b95b737",
			Name:   "My Test Container",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)
	client := mockclient.NewMockClient()
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{c}, nil)
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c.Name(), args.Get(0).(container.Container).Name())
			locker.Lock()
			if !touch {
				touch = true
				wg1.Done()
			}
			locker.Unlock()
		}).Return(nil)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		RunTickerTestContainers(ctx, client)
		wg2.Done()
	}()
	wg1.Wait()
	cancel()
	wg2.Wait()

	assert.True(t, touch)
	client.AssertExpectations(t)
}
