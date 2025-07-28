package types

import (
	"encoding/json"
	"fmt"
)

const (
	HostResourceType       = "host"
	HostReporterType       = "hbi"
	HostReporterInstanceID = "redhat.com"
	HostAPIHref            = "https://apiHref.com/"
	HostConsoleHref        = "https://www.console.com/"
	HostReporterVersion    = "1.0"
)

// HostMessage represents the structure of a Host Debezium change event
type HostMessage struct {
	Schema  map[string]interface{} `json:"schema"`
	Payload HostPayload            `json:"payload"`
}

type HostPayload struct {
	ID                    string     `json:"id"`
	AnsibleHost           string     `json:"ansible_host"`
	OrganizationID        string     `json:"organization_id"`
	InsightsID            string     `json:"insights_id"`
	SubscriptionManagerID string     `json:"subscription_manager_id"`
	SatelliteID           string     `json:"satellite_id"`
	Groups                GroupSlice `json:"groups"`
}

type Group struct {
	ID string `json:"id"`
}

// GroupSlice is a custom type that can unmarshal both JSON arrays and JSON strings containing arrays
type GroupSlice []Group

// UnmarshalJSON implements custom unmarshaling for GroupSlice
func (gs *GroupSlice) UnmarshalJSON(data []byte) error {
	// First try to unmarshal as a direct array
	var groups []Group
	if err := json.Unmarshal(data, &groups); err == nil {
		*gs = GroupSlice(groups)
		return nil
	}

	// If that fails, try to unmarshal as a string containing JSON
	var groupsStr string
	if err := json.Unmarshal(data, &groupsStr); err != nil {
		return fmt.Errorf("groups field is neither an array nor a string: %v", err)
	}

	// Parse the JSON string
	if err := json.Unmarshal([]byte(groupsStr), &groups); err != nil {
		return fmt.Errorf("failed to parse groups JSON string: %v", err)
	}

	*gs = GroupSlice(groups)
	return nil
}

type CanonicalFacts struct {
	BiosUUID              string `json:"bios_uuid"`
	InsightsID            string `json:"insights_id"`
	SubscriptionManagerID string `json:"subscription_manager_id"`
}
