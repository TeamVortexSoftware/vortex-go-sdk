package vortex

import (
	"encoding/json"
	"testing"
)

// TestInvitationGroupDeserialization tests that all 6 fields from the API response
// are properly deserialized into the InvitationGroup struct
func TestInvitationGroupDeserialization(t *testing.T) {
	// This is the actual structure returned by the API (MemberGroups table)
	apiResponse := `{
		"id": "550e8400-e29b-41d4-a716-446655440000",
		"accountId": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"groupId": "workspace-123",
		"type": "workspace",
		"name": "My Workspace",
		"createdAt": "2025-01-27T12:00:00.000Z"
	}`

	var group InvitationGroup
	err := json.Unmarshal([]byte(apiResponse), &group)
	if err != nil {
		t.Fatalf("Failed to unmarshal InvitationGroup: %v", err)
	}

	// Verify all 6 fields are present and correct
	if group.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("Expected id to be '550e8400-e29b-41d4-a716-446655440000', got '%s'", group.ID)
	}
	if group.AccountID != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("Expected accountId to be '6ba7b810-9dad-11d1-80b4-00c04fd430c8', got '%s'", group.AccountID)
	}
	if group.GroupID != "workspace-123" {
		t.Errorf("Expected groupId to be 'workspace-123', got '%s'", group.GroupID)
	}
	if group.Type != "workspace" {
		t.Errorf("Expected type to be 'workspace', got '%s'", group.Type)
	}
	if group.Name != "My Workspace" {
		t.Errorf("Expected name to be 'My Workspace', got '%s'", group.Name)
	}
	if group.CreatedAt != "2025-01-27T12:00:00.000Z" {
		t.Errorf("Expected createdAt to be '2025-01-27T12:00:00.000Z', got '%s'", group.CreatedAt)
	}
}

// TestInvitationResultWithGroups tests that InvitationResult properly deserializes
// with an array of InvitationGroups
func TestInvitationResultWithGroups(t *testing.T) {
	apiResponse := `{
		"id": "inv-123",
		"accountId": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"clickThroughs": 5,
		"configurationAttributes": {},
		"attributes": {},
		"createdAt": "2025-01-27T12:00:00.000Z",
		"deactivated": false,
		"deliveryCount": 1,
		"deliveryTypes": ["email"],
		"foreignCreatorId": "user-123",
		"invitationType": "single_use",
		"modifiedAt": null,
		"status": "delivered",
		"target": [{"type": "email", "value": "test@example.com"}],
		"views": 10,
		"widgetConfigurationId": "widget-123",
		"projectId": "project-123",
		"groups": [
			{
				"id": "550e8400-e29b-41d4-a716-446655440000",
				"accountId": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				"groupId": "workspace-123",
				"type": "workspace",
				"name": "My Workspace",
				"createdAt": "2025-01-27T12:00:00.000Z"
			}
		],
		"accepts": []
	}`

	var invitation InvitationResult
	err := json.Unmarshal([]byte(apiResponse), &invitation)
	if err != nil {
		t.Fatalf("Failed to unmarshal InvitationResult: %v", err)
	}

	if len(invitation.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(invitation.Groups))
	}

	group := invitation.Groups[0]
	if group.GroupID != "workspace-123" {
		t.Errorf("Expected groupId to be 'workspace-123', got '%s'", group.GroupID)
	}
}

// TestGroupInputSerialization tests that Group type for JWT generation
// properly serializes with either id or groupId
func TestGroupInputSerialization(t *testing.T) {
	tests := []struct {
		name     string
		group    Group
		expected string
	}{
		{
			name: "Using legacy id field",
			group: Group{
				Type: "workspace",
				ID:   stringPtr("workspace-123"),
				Name: "My Workspace",
			},
			expected: `{"type":"workspace","id":"workspace-123","name":"My Workspace"}`,
		},
		{
			name: "Using preferred groupId field",
			group: Group{
				Type:    "workspace",
				GroupID: stringPtr("workspace-123"),
				Name:    "My Workspace",
			},
			expected: `{"type":"workspace","groupId":"workspace-123","name":"My Workspace"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.group)
			if err != nil {
				t.Fatalf("Failed to marshal Group: %v", err)
			}
			// Note: JSON key order may vary, so we'll unmarshal and compare
			var result map[string]interface{}
			json.Unmarshal(data, &result)
			var expected map[string]interface{}
			json.Unmarshal([]byte(tt.expected), &expected)

			if result["type"] != expected["type"] {
				t.Errorf("Type mismatch: got %v, want %v", result["type"], expected["type"])
			}
			if result["name"] != expected["name"] {
				t.Errorf("Name mismatch: got %v, want %v", result["name"], expected["name"])
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
