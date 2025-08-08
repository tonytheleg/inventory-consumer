package consumer

import (
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"github.com/stretchr/testify/assert"
)

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name              string
		expectedOperation string
		expectedVersion   string
		expectedTxid      string
		msg               *kafka.Message
		expectErr         bool
	}{
		{
			name:              "Create Operation",
			expectedOperation: OperationTypeReportResource,
			expectedVersion:   defaultApiVersion,
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeReportResource)},
					{Key: "version", Value: []byte(defaultApiVersion)},
				},
			},
			expectErr: false,
		},
		{
			name:              "Update Operation",
			expectedOperation: OperationTypeReportResource,
			expectedVersion:   defaultApiVersion,
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeReportResource)},
					{Key: "version", Value: []byte(defaultApiVersion)},
				},
			},
			expectErr: false,
		},
		{
			name:              "Delete Operation",
			expectedOperation: OperationTypeDeleteResource,
			expectedVersion:   defaultApiVersion,
			expectedTxid:      "",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeDeleteResource)},
					{Key: "version", Value: []byte(defaultApiVersion)},
				},
			},
			expectErr: false,
		},
		{
			name:              "Missing Operation Header",
			expectedOperation: "",
			expectedVersion:   defaultApiVersion,
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "version", Value: []byte(defaultApiVersion)},
				},
			},
			expectErr: true,
		},
		{
			name:              "Missing Operation Value",
			expectedOperation: "",
			expectedVersion:   defaultApiVersion,
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte{}},
					{Key: "version", Value: []byte(defaultApiVersion)},
				},
			},
			expectErr: true,
		},
		{
			name:              "Missing Version Header",
			expectedOperation: OperationTypeReportResource,
			expectedVersion:   "",
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeReportResource)},
				},
			},
			expectErr: true,
		},
		{
			name:              "Missing Version Value",
			expectedOperation: OperationTypeReportResource,
			expectedVersion:   "",
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeReportResource)},
					{Key: "version", Value: []byte{}},
				},
			},
			expectErr: true,
		},
		{
			name:              "Extra Headers",
			expectedOperation: OperationTypeReportResource,
			expectedVersion:   defaultApiVersion,
			expectedTxid:      "123456",
			msg: &kafka.Message{
				Headers: []kafka.Header{
					{Key: "operation", Value: []byte(OperationTypeReportResource)},
					{Key: "version", Value: []byte(defaultApiVersion)},
					{Key: "unused-header", Value: []byte("unused-header-data")},
				},
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsedHeaders, err := ParseHeaders(test.msg)
			if test.expectErr {
				assert.NotNil(t, err)
				assert.Empty(t, parsedHeaders.Operation)
				assert.Empty(t, parsedHeaders.Version)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, parsedHeaders.Operation, test.expectedOperation)
				assert.Equal(t, parsedHeaders.Version, test.expectedVersion)
			}
		})
	}
}

func TestParseCreateOrUpdateMessage(t *testing.T) {
	expected := makeReportResourceRequest()
	var req v1beta2.ReportResourceRequest
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
	var req v1beta2.DeleteResourceRequest
	err := ParseDeleteMessage([]byte(testDeleteMessage), &req)
	assert.Nil(t, err)
	assert.Equal(t, expected.Reference.ResourceId, req.Reference.ResourceId)
	assert.Equal(t, expected.Reference.ResourceType, req.Reference.ResourceType)
	assert.True(t, reflect.DeepEqual(expected.Reference.Reporter, req.Reference.Reporter))

}
