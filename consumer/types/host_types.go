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

	// Error messages for GroupSlice unmarshaling
	ErrGroupsNotArrayOrString = "groups field is neither an array nor a string"
	ErrGroupsInvalidJSON      = "failed to parse groups JSON string"
)

// HostMessage represents the structure of a Host Debezium change event
type HostMessage struct {
	Schema  map[string]interface{} `json:"schema"`
	Payload HostPayload            `json:"payload"`
}

type HostPayload struct {
	ID                    string     `json:"id"`
	AnsibleHost           string     `json:"ansible_host"`
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
		return fmt.Errorf("%s: %v", ErrGroupsNotArrayOrString, err)
	}

	// Parse the JSON string
	if err := json.Unmarshal([]byte(groupsStr), &groups); err != nil {
		return fmt.Errorf("%s: %v", ErrGroupsInvalidJSON, err)
	}

	*gs = GroupSlice(groups)
	return nil
}

type CanonicalFacts struct {
	BiosUUID              string `json:"bios_uuid"`
	InsightsID            string `json:"insights_id"`
	SubscriptionManagerID string `json:"subscription_manager_id"`
}
