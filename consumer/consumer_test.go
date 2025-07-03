package consumer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/stretchr/testify/mock"

	"github.com/project-kessel/inventory-consumer/internal/mocks"
	metricscollector "github.com/project-kessel/inventory-consumer/metrics"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/stretchr/testify/assert"

	. "github.com/project-kessel/inventory-api/cmd/common"
)

const (
	testMessageKey            = `{"schema":{"type":"string","optional":false},"payload":"00000000-0000-0000-0000-000000000000"}`
	testCreateOrUpdateMessage = `{"schema":{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"},{"type":"string","optional":true,"field":"reporter_type"},{"type":"string","optional":true,"field":"reporter_instance_id"},{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"local_resource_id"},{"type":"string","optional":true,"field":"api_href"},{"type":"string","optional":true,"field":"console_href"},{"type":"string","optional":true,"field":"reporter_version"}],"optional":true,"name":"metadata"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"workspace_id"}],"optional":true,"name":"common"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"satellite_id"},{"type":"string","optional":true,"field":"subscription_manager_id"},{"type":"string","optional":true,"field":"insights_inventory_id"},{"type":"string","optional":true,"field":"ansible_host"}],"optional":true,"name":"reporter"}],"optional":true,"name":"representations"}],"optional":true,"name":"payload"},"payload":{"type":"host","reporter_type":"hbi","reporter_instance_id":"00000000-0000-0000-0000-000000000000","representations":{"metadata":{"local_resource_id":"00000000-0000-0000-0000-000000000000","api_href":"https://apiHref.com/","console_href":"https://www.console.com/","reporter_version":"2.7.16"},"common":{"workspace_id":"00000000-0000-0000-0000-000000000000"},"reporter":{"satellite_id":"00000000-0000-0000-0000-000000000000","subscription_manager_id":"00000000-0000-0000-0000-000000000000","insights_inventory_id":"00000000-0000-0000-0000-000000000000","ansible_host":"my-ansible-host"}}}}`
	testDeleteMessage         = `{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"resource_type"},{"type":"string","optional":true,"field":"resource_id"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"}],"optional":true,"name":"reporter"}],"optional":true,"name":"reference"}],"optional":true,"name":"payload"},"payload":{"reference":{"resource_type":"host","resource_id":"00000000-0000-0000-0000-000000000000","reporter":{"type":"hbi"}}}}`
)

type TestCase struct {
	name            string
	description     string
	options         *Options
	config          *Config
	completedConfig CompletedConfig
	inv             InventoryConsumer
	metrics         metricscollector.MetricsCollector
	logger          *log.Helper
}

// TestSetup creates a test struct that calls most of the initial constructor methods we intend to test in unit tests.
func (t *TestCase) TestSetup() []error {
	t.options = NewOptions()
	t.options.BootstrapServers = []string{"localhost:9092"}
	t.options.Topics = []string{"test-topic"}
	t.config = NewConfig(t.options)
	t.config.AuthConfig.Enabled = false

	_, logger := InitLogger("info", LoggerOptions{})
	t.logger = log.NewHelper(log.With(logger, "subsystem", "inventoryConsumer"))

	var errs []error
	var err error

	if errList := t.options.Complete(); errList != nil {
		errs = append(errs, errList...)
	}
	if errList := t.options.Validate(); errList != nil {
		errs = append(errs, errList...)
	}
	cfg, errList := NewConfig(t.options).Complete()
	t.completedConfig = cfg
	if errList != nil {
		errs = append(errs, errList...)
	}

	consumer := mocks.MockConsumer{}
	t.inv, err = New(t.completedConfig, nil, t.logger)
	if err != nil {
		errs = append(errs, err)
	}

	t.inv.Consumer = &consumer

	err = t.metrics.New(t.config.Topics)
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

func makeReportResourceRequest() kesselv2.ReportResourceRequest {
	commonData, _ := structpb.NewStruct(map[string]interface{}{
		"workspace_id": "00000000-0000-0000-0000-000000000000",
	})

	reporterData, _ := structpb.NewStruct(map[string]interface{}{
		"satellite_id":            "00000000-0000-0000-0000-000000000000",
		"subscription_manager_id": "00000000-0000-0000-0000-000000000000",
		"insights_inventory_id":   "00000000-0000-0000-0000-000000000000",
		"ansible_host":            "my-ansible-host",
	})

	return kesselv2.ReportResourceRequest{
		Type:               "host",
		ReporterType:       "hbi",
		ReporterInstanceId: "00000000-0000-0000-0000-000000000000",
		Representations: &kesselv2.ResourceRepresentations{
			Metadata: &kesselv2.RepresentationMetadata{
				LocalResourceId: "00000000-0000-0000-0000-000000000000",
				ApiHref:         "https://apiHref.com/",
				ConsoleHref:     ToPointer("https://www.console.com/"),
				ReporterVersion: ToPointer("2.7.16"),
			},
			Common:   commonData,
			Reporter: reporterData,
		},
		WriteVisibility: 0,
	}
}

func makeDeleteResourceRequest() kesselv2.DeleteResourceRequest {
	return kesselv2.DeleteResourceRequest{
		Reference: &kesselv2.ResourceReference{
			ResourceType: "host",
			ResourceId:   "00000000-0000-0000-0000-000000000000",
			Reporter: &kesselv2.ReporterReference{
				Type: "hbi",
			},
		},
	}
}

func TestNewConsumerSetup(t *testing.T) {
	test := TestCase{
		name:        "TestNewConsumerSetup",
		description: "ensures setting up a new consumer, including options and configs functions",
	}
	errs := test.TestSetup()
	assert.Nil(t, errs)
}

func TestParseCreateOrUpdateMessage(t *testing.T) {
	expected := makeReportResourceRequest()
	var req kesselv2.ReportResourceRequest
	err := ParseCreateOrUpdateMessage([]byte(testCreateOrUpdateMessage), &req)
	assert.Nil(t, err)
	assert.Equal(t, expected.InventoryId, req.InventoryId)
	assert.Equal(t, expected.Type, req.Type)
	assert.Equal(t, expected.ReporterType, req.ReporterType)
	assert.Equal(t, expected.ReporterInstanceId, req.ReporterInstanceId)
	assert.True(t, reflect.DeepEqual(expected.Representations.Metadata, req.Representations.Metadata))
	assert.True(t, reflect.DeepEqual(expected.Representations.Common.AsMap(), req.Representations.Common.AsMap()))
	assert.True(t, reflect.DeepEqual(expected.Representations.Reporter.AsMap(), req.Representations.Reporter.AsMap()))
}

func TestParseDeleteMessage(t *testing.T) {
	expected := makeDeleteResourceRequest()
	var req kesselv2.DeleteResourceRequest
	err := ParseDeleteMessage([]byte(testDeleteMessage), &req)
	assert.Nil(t, err)
	assert.Equal(t, expected.Reference.ResourceId, req.Reference.ResourceId)
	assert.Equal(t, expected.Reference.ResourceType, req.Reference.ResourceType)
	assert.True(t, reflect.DeepEqual(expected.Reference.Reporter, req.Reference.Reporter))

}

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name              string
		expectedOperation string
		expectedTxid      string
		msg               *kafka.Message
		expectErr         bool
	}{
		{
			name:              "Create Operation",
			expectedOperation: OperationTypeCreated,
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeCreated)},
				},
			},
			expectErr: false,
		},
		{
			name:              "Update Operation",
			expectedOperation: OperationTypeUpdated,
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeUpdated)},
				},
			},
			expectErr: false,
		},
		{
			name:              "Delete Operation",
			expectedOperation: OperationTypeDeleted,
			expectedTxid:      "",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeDeleted)},
				},
			},
			expectErr: false,
		},
		{
			name:              "Missing Operation Header",
			expectedOperation: "",
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{},
			},
			expectErr: true,
		},
		{
			name:              "Missing Operation Value",
			expectedOperation: "",
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte{}},
				},
			},
			expectErr: true,
		},
		{
			name:              "Extra Headers",
			expectedOperation: OperationTypeCreated,
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeCreated)},
					{Key: "unused-header", Value: []byte("unused-header-data")},
				},
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedHeaders, err := ParseHeaders(test.msg, requiredHeaders)
			if test.expectErr {
				assert.NotNil(t, err)
				assert.Nil(t, parsedHeaders)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, parsedHeaders["operation"], test.expectedOperation)
				assert.LessOrEqual(t, len(parsedHeaders), 2)
			}
		})
	}
}

func TestInventoryConsumer_Retry(t *testing.T) {
	tests := []struct {
		description    string
		funcToExecute  func() (interface{}, error)
		expectedResult interface{}
		expectedErr    error
	}{
		{
			description:    "retry returns no error after executing function",
			funcToExecute:  func() (interface{}, error) { return "success", nil },
			expectedResult: "success",
			expectedErr:    nil,
		},
		{
			description:    "retry fails and returns MaxRetriesError",
			funcToExecute:  func() (interface{}, error) { return "fail", ErrMaxRetries },
			expectedResult: nil,
			expectedErr:    ErrMaxRetries,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			tester := TestCase{
				name:        "TestInventoryConsumer-Retry",
				description: test.description,
			}
			errs := tester.TestSetup()
			assert.Nil(t, errs)

			result, err := tester.inv.Retry(test.funcToExecute)
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestInventoryConsumer_ProcessMessage(t *testing.T) {
	tests := []struct {
		name              string
		expectedOperation string
		msg               *kafka.Message
		clientEnabled     bool
	}{
		{
			name:              "Create Operation",
			expectedOperation: OperationTypeCreated,
			msg: &kafka.Message{
				Key:   []byte(testMessageKey),
				Value: []byte(testCreateOrUpdateMessage),
			},
			clientEnabled: true,
		},
		{
			name:              "Update Operation",
			expectedOperation: OperationTypeUpdated,
			msg: &kafka.Message{
				Key:   []byte(testMessageKey),
				Value: []byte(testCreateOrUpdateMessage),
			},
			clientEnabled: true,
		},
		{
			name:              "Delete Operation",
			expectedOperation: OperationTypeDeleted,
			msg: &kafka.Message{
				Key:   []byte(testMessageKey),
				Value: []byte(testDeleteMessage),
			},
			clientEnabled: true,
		},
		{
			name:              "Fake Operation",
			expectedOperation: "fake-operation",
			msg:               &kafka.Message{},
			clientEnabled:     true,
		},
		{
			name:              "Created but inventory client disabled",
			expectedOperation: OperationTypeCreated,
			msg: &kafka.Message{
				Key:   []byte(testMessageKey),
				Value: []byte(testDeleteMessage),
			},
			clientEnabled: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tester := TestCase{}
			errs := tester.TestSetup()
			assert.Nil(t, errs)

			client := &mocks.MockClient{}
			client.On("CreateOrUpdateResource", mock.Anything).Return(&kesselv2.ReportResourceResponse{}, nil)
			client.On("DeleteResource", mock.Anything).Return(&kesselv2.DeleteResourceResponse{}, nil)
			client.On("IsEnabled").Return(test.clientEnabled)
			tester.inv.Client = client

			headers := []kafka.Header{
				{Key: "operation", Value: []byte(test.expectedOperation)},
			}
			test.msg.Headers = headers
			parsedHeaders, err := ParseHeaders(test.msg, requiredHeaders)
			operation := parsedHeaders["operation"]
			assert.Nil(t, err)
			assert.Equal(t, parsedHeaders["operation"], test.expectedOperation)

			if (test.expectedOperation == OperationTypeCreated) || test.expectedOperation == OperationTypeUpdated && test.clientEnabled {
				err := tester.inv.ProcessMessage(operation, test.msg)
				assert.Nil(t, err)
			} else {
				err := tester.inv.ProcessMessage(operation, test.msg)
				assert.Nil(t, err)
			}
		})
	}
}
func TestCheckIfCommit(t *testing.T) {
	tests := []struct {
		name      string
		partition kafka.TopicPartition
		expected  bool
	}{
		{
			name: "modulus of the partition offset equates to true",
			partition: kafka.TopicPartition{
				Offset: kafka.Offset(10),
			},
			expected: true,
		},
		{
			name: "modulus of the partition offset does not equate to true",
			partition: kafka.TopicPartition{
				Offset: kafka.Offset(1),
			},
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := CheckIfCommit(test.partition)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestCommitStoredOffsets(t *testing.T) {
	tests := []struct {
		name                                      string
		storedOffsets, response, remainingOffsets []kafka.TopicPartition
		responseErr                               error
	}{
		{
			name: "single stored offset is committed without error",
			storedOffsets: []kafka.TopicPartition{
				{Offset: kafka.Offset(10), Partition: 0},
			},
			response: []kafka.TopicPartition{
				{Offset: kafka.Offset(10), Partition: 0},
			},
			remainingOffsets: nil,
			responseErr:      nil,
		},
		{
			name: "all stored offsets are committed without error",
			storedOffsets: []kafka.TopicPartition{
				{Offset: kafka.Offset(10), Partition: 0},
				{Offset: kafka.Offset(11), Partition: 0},
				{Offset: kafka.Offset(1), Partition: 1},
				{Offset: kafka.Offset(2), Partition: 1},
				{Offset: kafka.Offset(12), Partition: 0},
				{Offset: kafka.Offset(13), Partition: 0},
				{Offset: kafka.Offset(3), Partition: 1},
				{Offset: kafka.Offset(4), Partition: 1},
			},
			response: []kafka.TopicPartition{
				{Offset: kafka.Offset(10), Partition: 0},
				{Offset: kafka.Offset(11), Partition: 0},
				{Offset: kafka.Offset(1), Partition: 1},
				{Offset: kafka.Offset(2), Partition: 1},
				{Offset: kafka.Offset(12), Partition: 0},
				{Offset: kafka.Offset(13), Partition: 0},
				{Offset: kafka.Offset(3), Partition: 1},
				{Offset: kafka.Offset(4), Partition: 1},
			},
			remainingOffsets: nil,
			responseErr:      nil,
		},
		{
			name: "Consumer.CommitOffsets returns error; offset storage is not cleared",
			storedOffsets: []kafka.TopicPartition{
				{Offset: kafka.Offset(10), Partition: 1},
			},
			response:         nil,
			remainingOffsets: []kafka.TopicPartition{{Offset: kafka.Offset(10), Partition: 1}},
			responseErr:      errors.New("commit failed"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tester := TestCase{}
			errs := tester.TestSetup()
			assert.Nil(t, errs)

			c := &mocks.MockConsumer{}
			c.On("CommitOffsets", mock.Anything).Return(test.response, test.responseErr)
			tester.inv.Consumer = c
			tester.inv.OffsetStorage = test.storedOffsets

			err := tester.inv.CommitStoredOffsets()
			assert.Equal(t, err, test.responseErr)
			assert.Equal(t, len(tester.inv.OffsetStorage), len(test.remainingOffsets))
			assert.Equal(t, tester.inv.OffsetStorage, test.remainingOffsets)
		})
	}
}
