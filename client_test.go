package vortex

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got %s", client.apiKey)
	}

	if client.baseURL != defaultBaseURL {
		t.Errorf("Expected baseURL to be %s, got %s", defaultBaseURL, client.baseURL)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customURL := "https://custom.example.com"
	client := NewClientWithOptions("test-api-key", customURL, nil)

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got %s", client.apiKey)
	}

	if client.baseURL != customURL {
		t.Errorf("Expected baseURL to be %s, got %s", customURL, client.baseURL)
	}
}

func TestGenerateJWT(t *testing.T) {
	// Test with valid API key format (UUID 12345678-1234-1234-1234-123456789012 encoded in base64url)
	client := NewClient("VRTX.EjRWeBI0EjQSNBI0VniQEg.test-key")

	user := &User{
		ID:          "user-123",
		Email:       "test@example.com",
		AdminScopes: []string{"autoJoin"},
	}

	jwt, err := client.GenerateJWT(user, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if jwt == "" {
		t.Error("Expected non-empty JWT")
	}

	// JWT should have 3 parts separated by dots
	parts := len(jwt) > 0 && len(splitJWT(jwt)) == 3
	if !parts {
		t.Error("JWT should have 3 parts separated by dots")
	}
}

func TestGenerateJWT_WithExtra(t *testing.T) {
	client := NewClient("VRTX.EjRWeBI0EjQSNBI0VniQEg.test-key")

	user := &User{
		ID:    "user-123",
		Email: "test@example.com",
	}

	extra := map[string]interface{}{
		"role":       "admin",
		"department": "Engineering",
	}

	jwt, err := client.GenerateJWT(user, extra)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if jwt == "" {
		t.Error("Expected non-empty JWT")
	}

	// JWT should have 3 parts separated by dots
	parts := len(jwt) > 0 && len(splitJWT(jwt)) == 3
	if !parts {
		t.Error("JWT should have 3 parts separated by dots")
	}
}

func TestGenerateJWT_InvalidAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
	}{
		{"missing parts", "invalid-key"},
		{"wrong prefix", "WRONG.EjRWeBI0EjQSNBI0VniQEg.test-key"},
		{"invalid encoding", "VRTX.invalid-base64.test-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiKey)
			user := &User{
				ID:    "user-123",
				Email: "test@example.com",
			}

			_, err := client.GenerateJWT(user, nil)
			if err == nil {
				t.Error("Expected error for invalid API key")
			}
		})
	}
}

func TestAPIRequest_Success(t *testing.T) {
	// Create mock server
	mockResponse := InvitationsResponse{
		Invitations: []InvitationResult{
			{
				ID:     "test-invitation-1",
				Status: "pending",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("Expected x-api-key header to be 'test-api-key', got %s", r.Header.Get("x-api-key"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClientWithOptions("test-api-key", server.URL, nil)

	// Test GetInvitationsByTarget
	invitations, err := client.GetInvitationsByTarget("email", "test@example.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(invitations) != 1 {
		t.Errorf("Expected 1 invitation, got %d", len(invitations))
	}

	if invitations[0].ID != "test-invitation-1" {
		t.Errorf("Expected invitation ID to be 'test-invitation-1', got %s", invitations[0].ID)
	}
}

func TestAPIRequest_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewClientWithOptions("test-api-key", server.URL, nil)

	// Test GetInvitationsByTarget
	_, err := client.GetInvitationsByTarget("email", "test@example.com")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", apiErr.StatusCode)
	}
}

func TestGetInvitation(t *testing.T) {
	mockInvitation := InvitationResult{
		ID:     "test-invitation-1",
		Status: "pending",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/invitations/test-invitation-1" {
			t.Errorf("Expected path '/api/v1/invitations/test-invitation-1', got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockInvitation)
	}))
	defer server.Close()

	client := NewClientWithOptions("test-api-key", server.URL, nil)

	invitation, err := client.GetInvitation("test-invitation-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if invitation.ID != "test-invitation-1" {
		t.Errorf("Expected invitation ID to be 'test-invitation-1', got %s", invitation.ID)
	}
}

func TestRevokeInvitation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/invitations/test-invitation-1" {
			t.Errorf("Expected path '/api/v1/invitations/test-invitation-1', got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClientWithOptions("test-api-key", server.URL, nil)

	err := client.RevokeInvitation("test-invitation-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAcceptInvitations(t *testing.T) {
	mockResult := InvitationResult{
		ID:     "accepted-invitation",
		Status: "accepted",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/invitations/accept" {
			t.Errorf("Expected path '/api/v1/invitations/accept', got %s", r.URL.Path)
		}

		// Verify request body
		var req AcceptInvitationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if len(req.InvitationIDs) != 2 {
			t.Errorf("Expected 2 invitation IDs, got %d", len(req.InvitationIDs))
		}

		if req.Target.Type != "email" {
			t.Errorf("Expected target type 'email', got %s", req.Target.Type)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResult)
	}))
	defer server.Close()

	client := NewClientWithOptions("test-api-key", server.URL, nil)

	target := InvitationTarget{
		Type:  "email",
		Value: "test@example.com",
	}

	result, err := client.AcceptInvitations([]string{"inv1", "inv2"}, target)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != "accepted-invitation" {
		t.Errorf("Expected result ID to be 'accepted-invitation', got %s", result.ID)
	}
}

// Helper functions for tests
func stringPtr(s string) *string {
	return &s
}

func splitJWT(jwt string) []string {
	parts := []string{}
	current := ""

	for _, char := range jwt {
		if char == '.' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}