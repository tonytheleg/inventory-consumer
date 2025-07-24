package transforms

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/project-kessel/inventory-consumer/consumer/types"
)

// TransformHostToReportResourceRequest transforms a Debezium message into a kesselv2.ReportResourceRequest
func TransformHostToReportResourceRequest(msg []byte) (*kesselv2.ReportResourceRequest, error) {
	var hostMsg types.HostMessage
	err := json.Unmarshal(msg, &hostMsg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling Debezium message: %v", err)
	}

	// Parse canonical_facts JSON field
	var canonicalFacts types.CanonicalFacts
	if hostMsg.Payload.CanonicalFacts != "" {
		err := json.Unmarshal([]byte(hostMsg.Payload.CanonicalFacts), &canonicalFacts)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling canonical_facts: %v", err)
		}
	}

	// Create a simplified structure that matches the expected format
	// First convert to the intermediate JSON structure
	intermediatePayload := map[string]interface{}{
		"type":                 types.HostResourceType,
		"reporter_type":        types.HostReporterType,
		"reporter_instance_id": hostMsg.Payload.ID,
		"representations": map[string]interface{}{
			"metadata": map[string]interface{}{
				"local_resource_id": hostMsg.Payload.ID,
				"api_href":          "https://apiHref.com/",
				"console_href":      "https://www.console.com/",
				"reporter_version":  "1.0",
			},
			"common": map[string]interface{}{
				"workspace_id": hostMsg.Payload.OrganizationID,
			},
			"reporter": map[string]interface{}{
				"satellite_id":   uuid.New().String(), // TODO: hook this up to the satellite id
				"sub_manager_id": uuid.New().String(), // TODO: hook this up to the sub manager id
				// "sub_manager_id":        canonicalFacts.SubscriptionManagerID,
				"insights_inventory_id": uuid.New().String(), // TODO: hook this up to the insights inventory id
				// "insights_inventory_id": hostMsg.Payload.ID,
				"ansible_host": hostMsg.Payload.AnsibleHost,
			},
			"common_resource_data": map[string]interface{}{
				"workspace_id": hostMsg.Payload.OrganizationID,
			},
		},
	}

	// Marshal and unmarshal to convert to the expected type
	payloadBytes, err := json.Marshal(intermediatePayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling intermediate payload: %v", err)
	}

	var request kesselv2.ReportResourceRequest
	err = json.Unmarshal(payloadBytes, &request)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling to ReportResourceRequest: %v", err)
	}

	return &request, nil
}

// TransformHostToDeleteResourceRequest transforms a tombstone message into a kesselv2.DeleteResourceRequest
// Extracts the resource ID from the message key since tombstones have empty values
func TransformHostToDeleteResourceRequest(msgValue []byte, msgKey []byte) (*kesselv2.DeleteResourceRequest, error) {
	// Extract ID from the key
	if len(msgKey) == 0 {
		return nil, fmt.Errorf("tombstone message has no key to extract resource ID")
	}

	var keyPayload struct {
		Payload struct {
			ID string `json:"id"`
		} `json:"payload"`
	}

	err := json.Unmarshal(msgKey, &keyPayload)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling message key for tombstone: %v", err)
	}

	resourceID := keyPayload.Payload.ID
	if resourceID == "" {
		return nil, fmt.Errorf("cannot extract resource ID from tombstone message key")
	}

	return &kesselv2.DeleteResourceRequest{
		Reference: &kesselv2.ResourceReference{
			ResourceType: types.HostResourceType,
			ResourceId:   resourceID,
			Reporter: &kesselv2.ReporterReference{
				Type: types.HostReporterType,
			},
		},
	}, nil
}

// IsHostDeleted checks if a Debezium message is a tombstone event (indicating deletion)
func IsHostDeleted(msgValue []byte) (bool, error) {
	// Tombstone events have empty/null values
	return len(msgValue) == 0 || isEmptyJSON(msgValue), nil
}

// isEmptyJSON checks if the byte slice represents empty JSON (null, whitespace only, etc.)
func isEmptyJSON(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	return trimmed == "" || trimmed == "null"
}
