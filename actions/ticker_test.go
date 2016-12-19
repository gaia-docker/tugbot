package actions

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gaia-docker/tugbot/container"
	"github.com/gaia-docker/tugbot/container/mockclient"
	"github.com/samalba/dockerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/kubernetes/staging/src/k8s.io/client-go/1.4/_vendor/golang.org/x/net/context"

	"errors"
	"sync"
	"testing"
	"time"
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
		RunTickerTestContainers(ctx, client, time.Second*10)
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
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).Return([]container.Container{c}, nil).Once()
	client.On("Inspect", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c.ID(), args.Get(0).(string))
		}).Return(&c, nil).Once()
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c.Name(), args.Get(0).(container.Container).Name())
			locker.Lock()
			if !touch {
				touch = true
				wg1.Done()
			}
			locker.Unlock()
		}).Return(nil).Once()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		RunTickerTestContainers(ctx, client, time.Second*10)
		wg2.Done()
	}()
	wg1.Wait()
	cancel()
	wg2.Wait()

	assert.True(t, touch)
	client.AssertExpectations(t)
}

func TestRunTickerTestContainers_Iteration2ContainsDifferentListContainers(t *testing.T) {
	var wg1, wg2 sync.WaitGroup
	wg1.Add(1)
	wg2.Add(1)
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			container.TugbotTest:       "true",
			container.TugbotEventTimer: "10s",
		},
	}
	c1 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Id:     "c1id",
			Name:   "c1",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)
	c2 := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Id:     "c2id",
			Name:   "c2",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)

	client := mockclient.NewMockClient()

	// Iteration 1 - c1
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).
		Return([]container.Container{c1}, nil).Once()
	client.On("Inspect", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c1.ID(), args.Get(0).(string))
		}).Return(&c1, nil).Once()
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			name := args.Get(0).(container.Container).Name()
			log.Info("Running container ", name)
			assert.Equal(t, c1.Name(), name)
		}).Return(nil).Once()

	// Iteration 2 - c2
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).
		Return([]container.Container{c2}, nil).Once()
	client.On("Inspect", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c2.ID(), args.Get(0).(string))
		}).Return(&c2, nil).Once()
	client.On("StartContainerFrom", mock.AnythingOfType("container.Container")).
		Run(func(args mock.Arguments) {
			name := args.Get(0).(container.Container).Name()
			log.Info("Running container ", name)
			assert.Equal(t, c2.Name(), name)
			wg1.Done()
		}).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		RunTickerTestContainers(ctx, client, time.Nanosecond*1)
		wg2.Done()
	}()
	log.Info("Wating for finish running container ", c2.Name())
	wg1.Wait()
	cancel()
	log.Info("Wating for quiting ticker")
	wg2.Wait()

	client.AssertExpectations(t)
}

func TestRunTickerTestContainers_FailedToFindBasicContainer(t *testing.T) {
	var wg1, wg2 sync.WaitGroup
	wg1.Add(1)
	wg2.Add(1)
	cc := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			container.TugbotTest:       "true",
			container.TugbotEventTimer: "1ms",
		},
	}
	c := *container.NewContainer(
		&dockerclient.ContainerInfo{
			Id:     "cid",
			Name:   "API tests",
			Config: cc,
			State:  stateExited,
		},
		nil,
	)

	client := mockclient.NewMockClient()
	client.On("ListContainers", mock.AnythingOfType("container.Filter")).
		Return([]container.Container{c}, nil).Once()
	// failed to find basic container (to start from it a new container)
	client.On("Inspect", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			assert.Equal(t, c.ID(), args.Get(0).(string))
			wg1.Done()
		}).Return(&container.Container{}, errors.New("Expected :-)")).Once()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		RunTickerTestContainers(ctx, client, time.Second*10)
		wg2.Done()
	}()
	log.Info("Wating for inpect of ", c.Name())
	wg1.Wait()
	cancel()
	log.Info("Wating for quiting ticker")
	wg2.Wait()

	client.AssertExpectations(t)
}
