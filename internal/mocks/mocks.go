package mocks

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
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

func (m *MockClient) CreateOrUpdateResource(request *kesselv2.ReportResourceRequest) (*kesselv2.ReportResourceResponse, error) {
	args := m.Called(request)
	return args.Get(0).(*kesselv2.ReportResourceResponse), args.Error(1)
}

func (m *MockClient) DeleteResource(request *kesselv2.DeleteResourceRequest) (*kesselv2.DeleteResourceResponse, error) {
	args := m.Called(request)
	return args.Get(0).(*kesselv2.DeleteResourceResponse), args.Error(1)
}

func (m *MockClient) IsEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}
