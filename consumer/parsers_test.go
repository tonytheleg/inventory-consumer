package consumer

import (
	"reflect"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/stretchr/testify/assert"
)

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
