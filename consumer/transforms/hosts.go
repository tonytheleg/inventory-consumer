package transforms

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/project-kessel/inventory-consumer/consumer/types"
	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

// TransformHostToReportResourceRequest transforms a Debezium message into a kesselv2.ReportResourceRequest
func TransformHostToReportResourceRequest(msg []byte) (*v1beta2.ReportResourceRequest, error) {
	var hostMsg types.HostMessage
	err := json.Unmarshal(msg, &hostMsg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling Debezium message: %v", err)
	}

	// Create a simplified structure that matches the expected format
	// First convert to the intermediate JSON structure
	intermediatePayload := map[string]interface{}{
		"type":                 types.HostResourceType,
		"reporter_type":        types.HostReporterType,
		"reporter_instance_id": types.HostReporterInstanceID,
		"representations": map[string]interface{}{
			"metadata": map[string]interface{}{
				"local_resource_id": hostMsg.Payload.ID,
				"api_href":          types.HostAPIHref,
				"console_href":      types.HostConsoleHref,
				"reporter_version":  types.HostReporterVersion,
			},
			"reporter": map[string]interface{}{
				"satellite_id":          hostMsg.Payload.SatelliteID,
				"sub_manager_id":        hostMsg.Payload.SubscriptionManagerID,
				"insights_inventory_id": hostMsg.Payload.InsightsID,
				"ansible_host":          hostMsg.Payload.AnsibleHost,
			},
			"common": map[string]interface{}{
				"workspace_id": hostMsg.Payload.Groups[0].ID,
			},
		},
	}

	// Marshal and unmarshal to convert to the expected type
	payloadBytes, err := json.Marshal(intermediatePayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling intermediate payload: %v", err)
	}

	var request v1beta2.ReportResourceRequest
	err = json.Unmarshal(payloadBytes, &request)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling to ReportResourceRequest: %v", err)
	}

	return &request, nil
}

// TransformHostToDeleteResourceRequest transforms a tombstone message into a kesselv2.DeleteResourceRequest
// Extracts the resource ID from the message key since tombstones have empty values
func TransformHostToDeleteResourceRequest(msgValue []byte, msgKey []byte) (*v1beta2.DeleteResourceRequest, error) {
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

	return &v1beta2.DeleteResourceRequest{
		Reference: &v1beta2.ResourceReference{
			ResourceType: types.HostResourceType,
			ResourceId:   resourceID,
			Reporter: &v1beta2.ReporterReference{
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
