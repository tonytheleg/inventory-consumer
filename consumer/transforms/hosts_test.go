package transforms

import (
	"testing"

	"github.com/project-kessel/inventory-consumer/consumer/types"
	"github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
	"github.com/stretchr/testify/assert"
)

const (
	// Test UUIDs
	testHostID1      = "11111111-1111-1111-1111-111111111111"
	testSatelliteID1 = "22222222-2222-2222-2222-222222222222"
	testSubMgrID1    = "33333333-3333-3333-3333-333333333333"
	testInsightsID1  = "44444444-4444-4444-4444-444444444444"
	testWorkspaceID1 = "55555555-5555-5555-5555-555555555555"

	testHostID2      = "66666666-6666-6666-6666-666666666666"
	testSatelliteID2 = "77777777-7777-7777-7777-777777777777"
	testSubMgrID2    = "88888888-8888-8888-8888-888888888888"
	testInsightsID2  = "99999999-9999-9999-9999-999999999999"
	testWorkspaceID2 = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	testWorkspaceID3 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	testHostID3      = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	testSatelliteID3 = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	testSubMgrID3    = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	testInsightsID3  = "ffffffff-ffff-ffff-ffff-ffffffffffff"

	testDeletedHostID = "12345678-1234-1234-1234-123456789012"

	// Test ansible hosts
	testAnsibleHost1 = "test-ansible-host.example.com"
	testAnsibleHost2 = "multi-group-host.example.com"
	testAnsibleHost3 = "no-groups-host.example.com"

	// Error messages
	errUnmarshalingDebezium = "error unmarshaling Debezium message"
	errUnmarshalingKey      = "error unmarshaling message key for tombstone"
	errNoKeyForTombstone    = "tombstone message has no key to extract resource ID"
	errNoResourceID         = "cannot extract resource ID from tombstone message key"
	errIndexOutOfRange      = "runtime error: index out of range"

	// Test messages
	testHostMessageValid = `{
		"payload": {
			"id": "` + testHostID1 + `",
			"satellite_id": "` + testSatelliteID1 + `",
			"subscription_manager_id": "` + testSubMgrID1 + `",
			"insights_id": "` + testInsightsID1 + `",
			"ansible_host": "` + testAnsibleHost1 + `",
			"groups": [
				{
					"id": "` + testWorkspaceID1 + `"
				}
			]
		}
	}`

	testHostMessageMultipleGroups = `{
		"payload": {
			"id": "` + testHostID2 + `",
			"satellite_id": "` + testSatelliteID2 + `",
			"subscription_manager_id": "` + testSubMgrID2 + `",
			"insights_id": "` + testInsightsID2 + `",
			"ansible_host": "` + testAnsibleHost2 + `",
			"groups": [
				{
					"id": "` + testWorkspaceID2 + `"
				},
				{
					"id": "` + testWorkspaceID3 + `"
				}
			]
		}
	}`

	testHostMessageInvalidJSON = `{invalid json`

	testHostMessageEmptyPayload = `{
		"payload": {}
	}`

	testHostMessageNoGroups = `{
		"payload": {
			"id": "` + testHostID3 + `",
			"satellite_id": "` + testSatelliteID3 + `",
			"subscription_manager_id": "` + testSubMgrID3 + `",
			"insights_id": "` + testInsightsID3 + `",
			"ansible_host": "` + testAnsibleHost3 + `",
			"groups": []
		}
	}`

	testTombstoneKey = `{
		"payload": {
			"id": "` + testDeletedHostID + `"
		}
	}`

	testTombstoneKeyNoID = `{
		"payload": {}
	}`

	testTombstoneKeyInvalidJSON = `{invalid`
)

func TestTransformHostToReportResourceRequest(t *testing.T) {
	tests := []struct {
		name          string
		message       []byte
		expectError   bool
		errorContains string
		validate      func(*testing.T, *v1beta2.ReportResourceRequest)
	}{
		{
			name:        "valid host message transforms correctly",
			message:     []byte(testHostMessageValid),
			expectError: false,
			validate: func(t *testing.T, req *v1beta2.ReportResourceRequest) {
				assert.Equal(t, types.HostResourceType, req.Type)
				assert.Equal(t, types.HostReporterType, req.ReporterType)
				assert.Equal(t, types.HostReporterInstanceID, req.ReporterInstanceId)

				// Check metadata
				assert.NotNil(t, req.Representations)
				assert.NotNil(t, req.Representations.Metadata)
				assert.Equal(t, testHostID1, req.Representations.Metadata.LocalResourceId)
				assert.Equal(t, types.HostAPIHref, req.Representations.Metadata.ApiHref)
				assert.Equal(t, types.HostConsoleHref, *req.Representations.Metadata.ConsoleHref)
				assert.Equal(t, types.HostReporterVersion, *req.Representations.Metadata.ReporterVersion)

				// Check reporter data
				reporterMap := req.Representations.Reporter.AsMap()
				assert.Equal(t, testSatelliteID1, reporterMap["satellite_id"])
				assert.Equal(t, testSubMgrID1, reporterMap["sub_manager_id"])
				assert.Equal(t, testInsightsID1, reporterMap["insights_inventory_id"])
				assert.Equal(t, testAnsibleHost1, reporterMap["ansible_host"])

				// Check common data
				commonMap := req.Representations.Common.AsMap()
				assert.Equal(t, testWorkspaceID1, commonMap["workspace_id"])
			},
		},
		{
			name:        "host message with multiple groups uses first group",
			message:     []byte(testHostMessageMultipleGroups),
			expectError: false,
			validate: func(t *testing.T, req *v1beta2.ReportResourceRequest) {
				commonMap := req.Representations.Common.AsMap()
				assert.Equal(t, testWorkspaceID2, commonMap["workspace_id"])
			},
		},
		{
			name:          "invalid JSON returns error",
			message:       []byte(testHostMessageInvalidJSON),
			expectError:   true,
			errorContains: errUnmarshalingDebezium,
		},
		{
			name:          "empty payload transforms but may have issues",
			message:       []byte(testHostMessageEmptyPayload),
			expectError:   true,
			errorContains: errIndexOutOfRange,
		},
		{
			name:          "no groups causes panic",
			message:       []byte(testHostMessageNoGroups),
			expectError:   true,
			errorContains: errIndexOutOfRange,
		},
		{
			name:          "nil message returns error",
			message:       nil,
			expectError:   true,
			errorContains: errUnmarshalingDebezium,
		},
		{
			name:          "empty message returns error",
			message:       []byte{},
			expectError:   true,
			errorContains: errUnmarshalingDebezium,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Use defer to catch panics for tests that expect them
			if test.errorContains == errIndexOutOfRange {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic
						assert.Contains(t, r.(error).Error(), test.errorContains)
					} else {
						t.Errorf("expected panic but none occurred")
					}
				}()
			}

			req, err := TransformHostToReportResourceRequest(test.message)

			if test.expectError && test.errorContains != errIndexOutOfRange {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.errorContains)
				assert.Nil(t, req)
			} else if !test.expectError {
				assert.Nil(t, err)
				assert.NotNil(t, req)
				if test.validate != nil {
					test.validate(t, req)
				}
			}
		})
	}
}

func TestTransformHostToDeleteResourceRequest(t *testing.T) {
	tests := []struct {
		name          string
		msgValue      []byte
		msgKey        []byte
		expectError   bool
		errorContains string
		validate      func(*testing.T, *v1beta2.DeleteResourceRequest)
	}{
		{
			name:        "valid tombstone transforms correctly",
			msgValue:    []byte{},
			msgKey:      []byte(testTombstoneKey),
			expectError: false,
			validate: func(t *testing.T, req *v1beta2.DeleteResourceRequest) {
				assert.NotNil(t, req.Reference)
				assert.Equal(t, types.HostResourceType, req.Reference.ResourceType)
				assert.Equal(t, testDeletedHostID, req.Reference.ResourceId)
				assert.NotNil(t, req.Reference.Reporter)
				assert.Equal(t, types.HostReporterType, req.Reference.Reporter.Type)
			},
		},
		{
			name:          "empty key returns error",
			msgValue:      []byte{},
			msgKey:        []byte{},
			expectError:   true,
			errorContains: errNoKeyForTombstone,
		},
		{
			name:          "nil key returns error",
			msgValue:      []byte{},
			msgKey:        nil,
			expectError:   true,
			errorContains: errNoKeyForTombstone,
		},
		{
			name:          "invalid JSON in key returns error",
			msgValue:      []byte{},
			msgKey:        []byte(testTombstoneKeyInvalidJSON),
			expectError:   true,
			errorContains: errUnmarshalingKey,
		},
		{
			name:          "key without ID returns error",
			msgValue:      []byte{},
			msgKey:        []byte(testTombstoneKeyNoID),
			expectError:   true,
			errorContains: errNoResourceID,
		},
		{
			name:          "key with empty ID returns error",
			msgValue:      []byte{},
			msgKey:        []byte(`{"payload": {"id": ""}}`),
			expectError:   true,
			errorContains: errNoResourceID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := TransformHostToDeleteResourceRequest(test.msgValue, test.msgKey)

			if test.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), test.errorContains)
				assert.Nil(t, req)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, req)
				if test.validate != nil {
					test.validate(t, req)
				}
			}
		})
	}
}

func TestIsHostDeleted(t *testing.T) {
	tests := []struct {
		name        string
		msgValue    []byte
		expected    bool
		expectError bool
	}{
		{
			name:        "empty byte slice is tombstone",
			msgValue:    []byte{},
			expected:    true,
			expectError: false,
		},
		{
			name:        "nil is tombstone",
			msgValue:    nil,
			expected:    true,
			expectError: false,
		},
		{
			name:        "null string is tombstone",
			msgValue:    []byte("null"),
			expected:    true,
			expectError: false,
		},
		{
			name:        "valid JSON is not tombstone",
			msgValue:    []byte(`{"data": "value"}`),
			expected:    false,
			expectError: false,
		},
		{
			name:        "non-empty string is not tombstone",
			msgValue:    []byte("some data"),
			expected:    false,
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := IsHostDeleted(test.msgValue)

			if test.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestIsEmptyJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "empty byte slice",
			data:     []byte{},
			expected: true,
		},
		{
			name:     "valid JSON object",
			data:     []byte(`{"key": "value"}`),
			expected: false,
		},
		{
			name:     "valid JSON array",
			data:     []byte(`[1, 2, 3]`),
			expected: false,
		},
		{
			name:     "string value",
			data:     []byte(`"string"`),
			expected: false,
		},
		{
			name:     "number value",
			data:     []byte(`123`),
			expected: false,
		},
		{
			name:     "boolean value",
			data:     []byte(`true`),
			expected: false,
		},
		{
			name:     "null with extra chars",
			data:     []byte(`null123`),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isEmptyJSON(test.data)
			assert.Equal(t, test.expected, result)
		})
	}
}
