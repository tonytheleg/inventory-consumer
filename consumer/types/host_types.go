package types

// HostMessage represents the structure of a Host Debezium change event
type HostMessage struct {
	Schema  map[string]interface{} `json:"schema"`
	Payload HostPayload            `json:"payload"`
}

type HostPayload struct {
	ID             string `json:"id"`
	Account        string `json:"account"`
	Hostname       string `json:"hostname"`
	CreatedOn      string `json:"created_on"`
	ModifiedOn     string `json:"modified_on"`
	Tags           string `json:"tags"`
	CanonicalFacts string `json:"canonical_facts"`
	SystemProfile  string `json:"system_profile_facts"`
	AnsibleHost    string `json:"ansible_host"`
	Reporter       string `json:"reporter"`
	OrganizationID string `json:"organization_id"`
	Deleted        string `json:"__deleted"`
}

type CanonicalFacts struct {
	BiosUUID              string `json:"bios_uuid"`
	InsightsID            string `json:"insights_id"`
	SubscriptionManagerID string `json:"subscription_manager_id"`
}
