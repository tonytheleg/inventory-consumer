package mocks

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"github.com/stretchr/testify/mock"
)

type MockConsumer struct {
	mock.Mock
}

type MockClient struct {
	mock.Mock
}

func (m *MockConsumer) CommitOffsets(offsets []kafka.TopicPartition) ([]kafka.TopicPartition, error) {
	args := m.Called(offsets)
	return args.Get(0).([]kafka.TopicPartition), args.Error(1)
}

func (m *MockConsumer) SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) (err error) {
	args := m.Called(topics, rebalanceCb)
	return args.Error(0)
}

func (m *MockConsumer) Poll(timeoutMs int) (event kafka.Event) {
	args := m.Called(timeoutMs)
	return args.Get(0).(kafka.Event)
}

func (m *MockConsumer) IsClosed() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConsumer) AssignmentLost() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockClient) CreateOrUpdateResource(request *v1beta2.ReportResourceRequest) (*v1beta2.ReportResourceResponse, error) {
	args := m.Called(request)
	return args.Get(0).(*v1beta2.ReportResourceResponse), args.Error(1)
}

func (m *MockClient) DeleteResource(request *v1beta2.DeleteResourceRequest) (*v1beta2.DeleteResourceResponse, error) {
	args := m.Called(request)
	return args.Get(0).(*v1beta2.DeleteResourceResponse), args.Error(1)
}

func (m *MockClient) IsEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}
