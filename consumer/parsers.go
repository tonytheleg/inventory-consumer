package consumer

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// defines all required headers for message processing

// ParseHeaders parses the header values in a kafka event and returns them as a map
// It also verifies that all required headers are set
func ParseHeaders(msg *kafka.Message, requiredHeaders []string) (map[string]string, error) {
	headers := make(map[string]string)
	for _, v := range msg.Headers {
		// ignores any extra headers
		if slices.Contains(requiredHeaders, v.Key) {
			headers[v.Key] = string(v.Value)
		}
	}

	// ensures all required header keys are present after parsing, but only operation is required to have a value to process messages
	headerKeys := slices.Sorted(maps.Keys(headers))
	required := slices.Sorted(slices.Values(requiredHeaders))

	if !slices.Equal(headerKeys, required) || headers["operation"] == "" {
		return nil, fmt.Errorf("required headers are missing which would result in message processing failures: %+v", headers)
	}
	return headers, nil
}

// ParseCreateOrUpdateMessage parses a kafka event and converts the data into the specified create/update request data type passed
func ParseCreateOrUpdateMessage(msg []byte, output interface{}) error {
	var msgPayload *MessagePayload

	// msg value is expected to be a valid JSON body for the passed request type
	err := json.Unmarshal(msg, &msgPayload)
	if err != nil {
		return fmt.Errorf("error unmarshaling msgPayload: %v", err)
	}

	payloadJson, err := json.Marshal(msgPayload.ReportResourceRequest)
	if err != nil {
		return fmt.Errorf("error marshaling request payload: %v", err)
	}

	err = json.Unmarshal(payloadJson, &output)
	if err != nil {
		return fmt.Errorf("error unmarshaling request payload: %v", err)
	}
	return nil
}

// ParseDeleteMessage parses a kafka event and converts the data into the specified delete request data type passed
func ParseDeleteMessage(msg []byte, output interface{}) error {
	var msgPayload *MessagePayload

	// msg value is expected to be a valid JSON body for a single relation
	err := json.Unmarshal(msg, &msgPayload)
	if err != nil {
		return fmt.Errorf("error unmarshaling msgPayload: %v", err)
	}

	payloadJson, err := json.Marshal(msgPayload.ReportResourceRequest)
	if err != nil {
		return fmt.Errorf("error marshaling tuple payload: %v", err)
	}

	err = json.Unmarshal(payloadJson, &output)
	if err != nil {
		return fmt.Errorf("error unmarshaling tuple payload: %v", err)
	}
	return nil
}

// ParseMessageKey parses the kafka event key to capture the resource id
// This ensures all events for a specific resource id are processed by the same consumer
func ParseMessageKey(msg []byte) (string, error) {
	var msgPayload *KeyPayload

	// msg key is expected to be the inventory_id of a resource
	err := json.Unmarshal(msg, &msgPayload)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling msgPayload: %v", err)
	}
	return msgPayload.LocalResourceID, nil
}
