package mockclient

import (
	"time"

	"github.com/gaia-docker/tugbot/container"
	"github.com/stretchr/testify/mock"
)

// MockClient is mock implementation of container.Client which is a wrapper for Docker API.
type MockClient struct {
	mock.Mock
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) ListContainers(cf container.Filter) ([]container.Container, error) {
	args := m.Called(cf)
	return args.Get(0).([]container.Container), args.Error(1)
}

func (m *MockClient) StopContainer(c container.Container, timeout time.Duration) error {
	args := m.Called(c, timeout)
	return args.Error(0)
}

func (m *MockClient) StartContainerFrom(c container.Container) error {
	args := m.Called(c)
	return args.Error(0)
}
