package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupSlice_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedGroups GroupSlice
		expectError    bool
		errorContains  string
	}{
		{
			name:  "direct JSON array with single group",
			input: `[{"id": "group1"}]`,
			expectedGroups: GroupSlice{
				{ID: "group1"},
			},
			expectError: false,
		},
		{
			name:  "direct JSON array with multiple groups",
			input: `[{"id": "group1"}, {"id": "group2"}, {"id": "group3"}]`,
			expectedGroups: GroupSlice{
				{ID: "group1"},
				{ID: "group2"},
				{ID: "group3"},
			},
			expectError: false,
		},
		{
			name:           "empty JSON array",
			input:          `[]`,
			expectedGroups: GroupSlice{},
			expectError:    false,
		},
		{
			name:  "JSON string containing array with single group",
			input: `"[{\"id\": \"group1\"}]"`,
			expectedGroups: GroupSlice{
				{ID: "group1"},
			},
			expectError: false,
		},
		{
			name:  "JSON string containing array with multiple groups",
			input: `"[{\"id\": \"group1\"}, {\"id\": \"group2\"}]"`,
			expectedGroups: GroupSlice{
				{ID: "group1"},
				{ID: "group2"},
			},
			expectError: false,
		},
		{
			name:           "JSON string containing empty array",
			input:          `"[]"`,
			expectedGroups: GroupSlice{},
			expectError:    false,
		},
		{
			name:          "invalid JSON - not array or string",
			input:         `{"id": "group1"}`,
			expectError:   true,
			errorContains: ErrGroupsNotArrayOrString,
		},
		{
			name:           "null value becomes empty slice",
			input:          `null`,
			expectedGroups: nil,
			expectError:    false,
		},
		{
			name:          "invalid JSON - number",
			input:         `123`,
			expectError:   true,
			errorContains: ErrGroupsNotArrayOrString,
		},
		{
			name:          "invalid JSON - boolean",
			input:         `true`,
			expectError:   true,
			errorContains: ErrGroupsNotArrayOrString,
		},
		{
			name:          "JSON string with malformed array",
			input:         `"[{\"id\": \"group1\", invalid json]"`,
			expectError:   true,
			errorContains: ErrGroupsInvalidJSON,
		},
		{
			name:          "JSON string with invalid escape sequences",
			input:         `"[{\"id\": \"group1\"]"`,
			expectError:   true,
			errorContains: ErrGroupsInvalidJSON,
		},
		{
			name:          "empty string",
			input:         `""`,
			expectError:   true,
			errorContains: ErrGroupsInvalidJSON,
		},
		{
			name:          "JSON string containing non-array value",
			input:         `"not an array"`,
			expectError:   true,
			errorContains: ErrGroupsInvalidJSON,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var gs GroupSlice
			err := json.Unmarshal([]byte(test.input), &gs)

			if test.expectError {
				assert.NotNil(t, err)
				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedGroups, gs)
			}
		})
	}
}
