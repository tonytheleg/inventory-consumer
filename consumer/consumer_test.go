package consumer

import (
	"errors"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/stretchr/testify/mock"

	"github.com/project-kessel/inventory-consumer/internal/mocks"
	metricscollector "github.com/project-kessel/inventory-consumer/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/stretchr/testify/assert"

	. "github.com/project-kessel/inventory-api/cmd/common"
)

const (
	testMessageKey            = `{"schema":{"type":"string","optional":false},"payload":"00000000-0000-0000-0000-000000000000"}`
	testCreateOrUpdateMessage = `{"schema":{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"},{"type":"string","optional":true,"field":"reporter_type"},{"type":"string","optional":true,"field":"reporter_instance_id"},{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"local_resource_id"},{"type":"string","optional":true,"field":"api_href"},{"type":"string","optional":true,"field":"console_href"},{"type":"string","optional":true,"field":"reporter_version"}],"optional":true,"name":"metadata"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"workspace_id"}],"optional":true,"name":"common"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"satellite_id"},{"type":"string","optional":true,"field":"subscription_manager_id"},{"type":"string","optional":true,"field":"insights_inventory_id"},{"type":"string","optional":true,"field":"ansible_host"}],"optional":true,"name":"reporter"}],"optional":true,"name":"representations"}],"optional":true,"name":"payload"},"payload":{"type":"host","reporter_type":"hbi","reporter_instance_id":"00000000-0000-0000-0000-000000000000","representations":{"metadata":{"local_resource_id":"00000000-0000-0000-0000-000000000000","api_href":"https://apiHref.com/","console_href":"https://www.console.com/","reporter_version":"2.7.16"},"common":{"workspace_id":"00000000-0000-0000-0000-000000000000"},"reporter":{"satellite_id":"00000000-0000-0000-0000-000000000000","subscription_manager_id":"00000000-0000-0000-0000-000000000000","insights_inventory_id":"00000000-0000-0000-0000-000000000000","ansible_host":"my-ansible-host"}}}}`
	testDeleteMessage         = `{"schema":{"type":"struct","fields":[{"type":"struct","fields":[{"type":"string","optional":true,"field":"resource_type"},{"type":"string","optional":true,"field":"resource_id"},{"type":"struct","fields":[{"type":"string","optional":true,"field":"type"}],"optional":true,"name":"reporter"}],"optional":true,"name":"reference"}],"optional":true,"name":"payload"},"payload":{"reference":{"resource_type":"host","resource_id":"00000000-0000-0000-0000-000000000000","reporter":{"type":"hbi"}}}}`
	testMigrationMessage      = `{"schema":{"type":"struct","fields":[{"type":"string","optional":true,"field":"id"},{"type":"string","optional":true,"field":"ansible_host"},{"type":"string","optional":true,"field":"insights_id"},{"type":"string","optional":true,"field":"subscription_manager_id"},{"type":"string","optional":true,"field":"satellite_id"},{"type":"string","optional":true,"field":"groups"}],"optional":true,"name":"payload"},"payload":{"id":"00000000-0000-0000-0000-000000000000","ansible_host":"my-ansible-host","insights_id":"00000000-0000-0000-0000-000000000000","subscription_manager_id":"00000000-0000-0000-0000-000000000000","satellite_id":"00000000-0000-0000-0000-000000000000","groups":"[{\"id\":\"00000000-0000-0000-0000-000000000000\"}]"}}`
	testMigrationKey          = `{"payload":{"id":"00000000-0000-0000-0000-000000000000"}}`
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

	// Create mock consumer first
	mockConsumer := &mocks.MockConsumer{}

	// Use NewWithConsumer to avoid creating a real Kafka connection
	t.inv, err = NewWithConsumer(t.completedConfig, nil, t.logger, mockConsumer)
	if err != nil {
		errs = append(errs, err)
	}

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
		setupMock         func(*mocks.MockClient)
		expectError       bool
	}{
		{
			name:              "Create Operation",
			expectedOperation: OperationTypeCreated,
			msg: &kafka.Message{
				Key:   []byte(testMessageKey),
				Value: []byte(testCreateOrUpdateMessage),
			},
			clientEnabled: true,
			setupMock: func(client *mocks.MockClient) {
				client.On("CreateOrUpdateResource", mock.Anything).Return(&kesselv2.ReportResourceResponse{}, nil)
			},
		},
		{
			name:              "Update Operation",
			expectedOperation: OperationTypeUpdated,
			msg: &kafka.Message{
				Key:   []byte(testMessageKey),
				Value: []byte(testCreateOrUpdateMessage),
			},
			clientEnabled: true,
			setupMock: func(client *mocks.MockClient) {
				client.On("CreateOrUpdateResource", mock.Anything).Return(&kesselv2.ReportResourceResponse{}, nil)
			},
		},
		{
			name:              "Delete Operation",
			expectedOperation: OperationTypeDeleted,
			msg: &kafka.Message{
				Key:   []byte(testMessageKey),
				Value: []byte(testDeleteMessage),
			},
			clientEnabled: true,
			setupMock: func(client *mocks.MockClient) {
				client.On("DeleteResource", mock.Anything).Return(&kesselv2.DeleteResourceResponse{}, nil)
			},
		},
		{
			name:              "Fake Operation",
			expectedOperation: "fake-operation",
			msg:               &kafka.Message{},
			clientEnabled:     true,
			setupMock:         func(client *mocks.MockClient) {},
		},
		{
			name:              "Created but inventory client disabled",
			expectedOperation: OperationTypeCreated,
			msg: &kafka.Message{
				Key:   []byte(testMessageKey),
				Value: []byte(testDeleteMessage),
			},
			clientEnabled: false,
			setupMock:     func(client *mocks.MockClient) {},
		},
		{
			name:              "Migration Operation - Create/Update Host",
			expectedOperation: OperationTypeMigration,
			msg: &kafka.Message{
				Key:   []byte(testMigrationKey),
				Value: []byte(testMigrationMessage),
			},
			clientEnabled: true,
			setupMock: func(client *mocks.MockClient) {
				client.On("CreateOrUpdateResource", mock.Anything).Return(&kesselv2.ReportResourceResponse{}, nil)
			},
		},
		{
			name:              "Migration Operation - Delete Host (tombstone)",
			expectedOperation: OperationTypeMigration,
			msg: &kafka.Message{
				Key:   []byte(testMigrationKey),
				Value: []byte(""), // Empty value indicates tombstone/deletion
			},
			clientEnabled: true,
			setupMock: func(client *mocks.MockClient) {
				client.On("DeleteResource", mock.Anything).Return(&kesselv2.DeleteResourceResponse{}, nil)
			},
		},
		{
			name:              "Migration Operation - Delete Host NotFound (should drop message)",
			expectedOperation: OperationTypeMigration,
			msg: &kafka.Message{
				Key:   []byte(testMigrationKey),
				Value: []byte(""),
			},
			clientEnabled: true,
			setupMock: func(client *mocks.MockClient) {
				// Return NotFound error on first attempt, which should cause message to be dropped
				client.On("DeleteResource", mock.Anything).Return(&kesselv2.DeleteResourceResponse{}, status.Error(codes.NotFound, "resource not found"))
			},
		},
		{
			name:              "Migration Operation - Create/Update with Retry",
			expectedOperation: OperationTypeMigration,
			msg: &kafka.Message{
				Key:   []byte(testMigrationKey),
				Value: []byte(testMigrationMessage),
			},
			clientEnabled: true,
			setupMock: func(client *mocks.MockClient) {
				// Fail first attempt, succeed on second
				client.On("CreateOrUpdateResource", mock.Anything).Return(&kesselv2.ReportResourceResponse{}, errors.New("temporary error")).Once()
				client.On("CreateOrUpdateResource", mock.Anything).Return(&kesselv2.ReportResourceResponse{}, nil).Once()
			},
		},
		{
			name:              "Migration Operation - Delete with Retry",
			expectedOperation: OperationTypeMigration,
			msg: &kafka.Message{
				Key:   []byte(testMigrationKey),
				Value: []byte(""),
			},
			clientEnabled: true,
			setupMock: func(client *mocks.MockClient) {
				// Fail first attempt with non-NotFound error, succeed on second
				client.On("DeleteResource", mock.Anything).Return(&kesselv2.DeleteResourceResponse{}, errors.New("temporary error")).Once()
				client.On("DeleteResource", mock.Anything).Return(&kesselv2.DeleteResourceResponse{}, nil).Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tester := TestCase{}
			errs := tester.TestSetup()
			assert.Nil(t, errs)

			client := &mocks.MockClient{}

			client.On("IsEnabled").Return(test.clientEnabled).Maybe()

			// Run test-specific mock setup
			test.setupMock(client)

			tester.inv.Client = client

			headers := []kafka.Header{
				{Key: "operation", Value: []byte(test.expectedOperation)},
			}
			test.msg.Headers = headers
			parsedHeaders, err := ParseHeaders(test.msg, requiredHeaders)
			operation := parsedHeaders["operation"]
			assert.Nil(t, err)
			assert.Equal(t, parsedHeaders["operation"], test.expectedOperation)

			err = tester.inv.ProcessMessage(operation, test.msg)
			if test.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			// Verify all expected mock calls were made
			client.AssertExpectations(t)
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

func TestInventoryConsumer_RebalanceCallback(t *testing.T) {
	tests := []struct {
		name                     string
		event                    kafka.Event
		assignmentLost           bool
		commitOffsetsError       error
		expectedCommitOffsetCall bool
		expectedError            error
	}{
		{
			name: "RevokedPartitions with assignment lost calls CommitStoredOffsets",
			event: kafka.RevokedPartitions{
				Partitions: []kafka.TopicPartition{
					{Topic: ToPointer("test-topic"), Partition: 0, Offset: kafka.Offset(10)},
				},
			},
			assignmentLost:           true,
			commitOffsetsError:       nil,
			expectedCommitOffsetCall: true,
			expectedError:            nil,
		},
		{
			name: "RevokedPartitions without assignment lost still calls CommitStoredOffsets",
			event: kafka.RevokedPartitions{
				Partitions: []kafka.TopicPartition{
					{Topic: ToPointer("test-topic"), Partition: 0, Offset: kafka.Offset(10)},
				},
			},
			assignmentLost:           false,
			commitOffsetsError:       nil,
			expectedCommitOffsetCall: true,
			expectedError:            nil,
		},
		{
			name: "RevokedPartitions with CommitStoredOffsets error returns error",
			event: kafka.RevokedPartitions{
				Partitions: []kafka.TopicPartition{
					{Topic: ToPointer("test-topic"), Partition: 0, Offset: kafka.Offset(10)},
				},
			},
			assignmentLost:           true,
			commitOffsetsError:       errors.New("commit failed"),
			expectedCommitOffsetCall: true,
			expectedError:            errors.New("commit failed"),
		},
		{
			name: "AssignedPartitions does not call CommitStoredOffsets",
			event: kafka.AssignedPartitions{
				Partitions: []kafka.TopicPartition{
					{Topic: ToPointer("test-topic"), Partition: 0, Offset: kafka.Offset(10)},
				},
			},
			assignmentLost:           false,
			commitOffsetsError:       nil,
			expectedCommitOffsetCall: false,
			expectedError:            nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tester := TestCase{}
			errs := tester.TestSetup()
			assert.Nil(t, errs)

			// Mock the consumer methods
			mockConsumer := &mocks.MockConsumer{}

			// Only mock AssignmentLost for RevokedPartitions events
			if _, isRevoked := test.event.(kafka.RevokedPartitions); isRevoked {
				mockConsumer.On("AssignmentLost").Return(test.assignmentLost)
			}

			// Set up CommitOffsets mock based on whether we expect it to be called
			if test.expectedCommitOffsetCall {
				mockConsumer.On("CommitOffsets", mock.Anything).Return([]kafka.TopicPartition{}, test.commitOffsetsError)
			}

			tester.inv.Consumer = mockConsumer

			// Add some offsets to storage to simulate having stored offsets
			tester.inv.OffsetStorage = []kafka.TopicPartition{
				{Topic: ToPointer("test-topic"), Partition: 0, Offset: kafka.Offset(5)},
			}

			// Call the RebalanceCallback method
			err := tester.inv.RebalanceCallback(nil, test.event)

			// Assert expectations
			assert.Equal(t, test.expectedError, err)
			mockConsumer.AssertExpectations(t)
		})
	}
}

func TestFormatOffsets(t *testing.T) {
	tests := []struct {
		name     string
		offsets  []kafka.TopicPartition
		expected string
	}{
		{
			name:     "empty slice returns empty string",
			offsets:  []kafka.TopicPartition{},
			expected: "",
		},
		{
			name: "single partition formats correctly",
			offsets: []kafka.TopicPartition{
				{Partition: 0, Offset: kafka.Offset(10)},
			},
			expected: "[0:10]",
		},
		{
			name: "multiple partitions with same partition number",
			offsets: []kafka.TopicPartition{
				{Partition: 0, Offset: kafka.Offset(10)},
				{Partition: 0, Offset: kafka.Offset(11)},
				{Partition: 0, Offset: kafka.Offset(12)},
			},
			expected: "[0:10],[0:11],[0:12]",
		},
		{
			name: "multiple partitions with different partition numbers",
			offsets: []kafka.TopicPartition{
				{Partition: 0, Offset: kafka.Offset(10)},
				{Partition: 1, Offset: kafka.Offset(5)},
				{Partition: 2, Offset: kafka.Offset(100)},
			},
			expected: "[0:10],[1:5],[2:100]",
		},
		{
			name: "mixed partitions and offsets",
			offsets: []kafka.TopicPartition{
				{Partition: 3, Offset: kafka.Offset(0)},
				{Partition: 0, Offset: kafka.Offset(999)},
				{Partition: 1, Offset: kafka.Offset(1)},
			},
			expected: "[3:0],[0:999],[1:1]",
		},
		{
			name: "large offset values",
			offsets: []kafka.TopicPartition{
				{Partition: 10, Offset: kafka.Offset(1000000)},
				{Partition: 999, Offset: kafka.Offset(9999999999)},
			},
			expected: "[10:1000000],[999:9999999999]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FormatOffsets(test.offsets)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestInventoryConsumer_Shutdown(t *testing.T) {
	tests := []struct {
		name                     string
		consumerClosed           bool
		hasStoredOffsets         bool
		commitOffsetsError       error
		closeError               error
		expectedError            error
		expectedCommitOffsetCall bool
		expectedCloseCall        bool
	}{
		{
			name:                     "consumer not closed, has stored offsets, successful commit and close",
			consumerClosed:           false,
			hasStoredOffsets:         true,
			commitOffsetsError:       nil,
			closeError:               nil,
			expectedError:            ErrClosed,
			expectedCommitOffsetCall: true,
			expectedCloseCall:        true,
		},
		{
			name:                     "consumer not closed, has stored offsets, commit fails but close succeeds",
			consumerClosed:           false,
			hasStoredOffsets:         true,
			commitOffsetsError:       errors.New("commit failed"),
			closeError:               nil,
			expectedError:            ErrClosed,
			expectedCommitOffsetCall: true,
			expectedCloseCall:        true,
		},
		{
			name:                     "consumer not closed, has stored offsets, commit succeeds but close fails",
			consumerClosed:           false,
			hasStoredOffsets:         true,
			commitOffsetsError:       nil,
			closeError:               errors.New("close failed"),
			expectedError:            errors.New("close failed"),
			expectedCommitOffsetCall: true,
			expectedCloseCall:        true,
		},
		{
			name:                     "consumer not closed, no stored offsets, close succeeds",
			consumerClosed:           false,
			hasStoredOffsets:         false,
			commitOffsetsError:       nil,
			closeError:               nil,
			expectedError:            ErrClosed,
			expectedCommitOffsetCall: false,
			expectedCloseCall:        true,
		},
		{
			name:                     "consumer not closed, no stored offsets, close fails",
			consumerClosed:           false,
			hasStoredOffsets:         false,
			commitOffsetsError:       nil,
			closeError:               errors.New("close failed"),
			expectedError:            errors.New("close failed"),
			expectedCommitOffsetCall: false,
			expectedCloseCall:        true,
		},
		{
			name:                     "consumer already closed - returns ErrClosed immediately",
			consumerClosed:           true,
			hasStoredOffsets:         true, // Even with stored offsets, they won't be committed
			commitOffsetsError:       nil,
			closeError:               nil,
			expectedError:            ErrClosed,
			expectedCommitOffsetCall: false,
			expectedCloseCall:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tester := TestCase{}
			errs := tester.TestSetup()
			assert.Nil(t, errs)

			// Mock the consumer methods
			mockConsumer := &mocks.MockConsumer{}
			mockConsumer.On("IsClosed").Return(test.consumerClosed)

			// Set up CommitOffsets mock based on whether we expect it to be called
			if test.expectedCommitOffsetCall {
				mockConsumer.On("CommitOffsets", mock.Anything).Return([]kafka.TopicPartition{}, test.commitOffsetsError)
			}

			// Set up Close mock based on whether we expect it to be called
			if test.expectedCloseCall {
				mockConsumer.On("Close").Return(test.closeError)
			}

			tester.inv.Consumer = mockConsumer

			// Set up offset storage based on test requirements
			if test.hasStoredOffsets {
				tester.inv.OffsetStorage = []kafka.TopicPartition{
					{Topic: ToPointer("test-topic"), Partition: 0, Offset: kafka.Offset(5)},
					{Topic: ToPointer("test-topic"), Partition: 1, Offset: kafka.Offset(10)},
				}
			} else {
				tester.inv.OffsetStorage = []kafka.TopicPartition{}
			}

			// Call the Shutdown method
			err := tester.inv.Shutdown()

			// Assert expectations
			if test.expectedError != nil {
				if errors.Is(test.expectedError, ErrClosed) {
					assert.Equal(t, ErrClosed, err)
				} else {
					assert.Equal(t, test.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			mockConsumer.AssertExpectations(t)
		})
	}
}
