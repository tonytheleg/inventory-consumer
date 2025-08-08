package consumer

import (
	"encoding/json"
	"fmt"

	"errors"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/mitchellh/mapstructure"
)

// defines all required headers for message processing

// ParseHeaders parses the header values in a kafka event and returns them as an EventHeaders object
// It also verifies that all required headers are set
func ParseHeaders(msg *kafka.Message) (EventHeaders, error) {
	var errs []error
	var headers EventHeaders

	mapHeaders := make(map[string]interface{})
	for _, v := range msg.Headers {
		// ignores any extra headers
		if _, ok := requiredHeaders[v.Key]; ok {
			mapHeaders[v.Key] = string(v.Value)
		}
	}
	config := &mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &headers,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return EventHeaders{}, fmt.Errorf("error creating decoder: %w", err)
	}
	if err := decoder.Decode(mapHeaders); err != nil {
		return EventHeaders{}, fmt.Errorf("error decoding headers: %w", err)
	}
	// validate all header values are set and have valid values -- return all errors if multiple are found
	if _, ok := validOperations[headers.Operation]; !ok {
		errs = append(errs, fmt.Errorf("required header 'operation' is missing or invalid: operation='%s'", headers.Operation))
	}
	if _, ok := validApiVersions[headers.Version]; !ok {
		errs = append(errs, fmt.Errorf("required header 'version' is missing or invalid: version='%s'", headers.Version))
	}
	if errs != nil {
		return EventHeaders{}, errors.Join(errs...)
	}
	return headers, nil
}

// ParseCreateOrUpdateMessage parses a kafka event and converts the data into the specified create/update request data type passed
func ParseCreateOrUpdateMessage(msg []byte, output interface{}) error {
	var msgPayload *MessagePayload

	// msg value is expected to be a valid JSON body for the passed request type
	err := json.Unmarshal(msg, &msgPayload)
	if err != nil {
		return fmt.Errorf("error unmarshaling msgPayload: %w", err)
	}

	payloadJson, err := json.Marshal(msgPayload.RequestPayload)
	if err != nil {
		return fmt.Errorf("error marshaling request payload: %w", err)
	}

	err = json.Unmarshal(payloadJson, &output)
	if err != nil {
		return fmt.Errorf("error unmarshaling request payload: %w", err)
	}
	return nil
}

// ParseDeleteMessage parses a kafka event and converts the data into the specified delete request data type passed
func ParseDeleteMessage(msg []byte, output interface{}) error {
	var msgPayload *MessagePayload

	// msg value is expected to be a valid JSON body for a single relation
	err := json.Unmarshal(msg, &msgPayload)
	if err != nil {
		return fmt.Errorf("error unmarshaling msgPayload: %w", err)
	}

	payloadJson, err := json.Marshal(msgPayload.RequestPayload)
	if err != nil {
		return fmt.Errorf("error marshaling tuple payload: %w", err)
	}

	err = json.Unmarshal(payloadJson, &output)
	if err != nil {
		return fmt.Errorf("error unmarshaling tuple payload: %w", err)
	}
	return nil
}
