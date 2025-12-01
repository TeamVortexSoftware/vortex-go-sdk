package vortex

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultBaseURL = "https://api.vortexsoftware.com"
	userAgent      = "vortex-go-sdk/1.0.0"
)

// Client represents a Vortex API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Vortex client
func NewClient(apiKey string) *Client {
	baseURL := os.Getenv("VORTEX_API_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewClientWithOptions creates a new Vortex client with custom options
func NewClientWithOptions(apiKey, baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// GenerateJWT creates a JWT token with the given user data and optional extra properties
//
// The user parameter should contain the user's ID, email, and optional admin scopes.
// If adminScopes is provided, the full array will be included in the JWT payload.
// The extra parameter can contain additional properties to include in the JWT payload.
//
// Example:
//
//	user := &vortex.User{
//	    ID:          "user-123",
//	    Email:       "user@example.com",
//	    AdminScopes: []string{"autoJoin"},
//	}
//	jwt, err := client.GenerateJWT(user, nil)
//
// Example with extra properties:
//
//	extra := map[string]interface{}{
//	    "role":       "admin",
//	    "department": "Engineering",
//	}
//	jwt, err := client.GenerateJWT(user, extra)
func (c *Client) GenerateJWT(user *User, extra map[string]interface{}) (string, error) {
	// Parse API key: format is VRTX.base64encodedId.key
	parts := strings.Split(c.apiKey, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid API key format")
	}

	prefix := parts[0]
	encodedID := parts[1]
	key := parts[2]

	if prefix != "VRTX" {
		return "", fmt.Errorf("invalid API key prefix")
	}

	// Decode the UUID from base64url
	uuidBytes, err := base64.RawURLEncoding.DecodeString(encodedID)
	if err != nil {
		return "", fmt.Errorf("failed to decode API key ID: %w", err)
	}

	// Convert bytes to UUID string
	id, err := uuid.FromBytes(uuidBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse UUID from API key: %w", err)
	}

	// Step 1: Derive signing key from API key + ID
	signingKeyHmac := hmac.New(sha256.New, []byte(key))
	signingKeyHmac.Write([]byte(id.String()))
	signingKey := signingKeyHmac.Sum(nil)

	// Step 2: Build header + payload
	now := time.Now().Unix()
	expires := now + 3600 // 1 hour

	header := JWTHeader{
		IAT: now,
		Alg: "HS256",
		Typ: "JWT",
		Kid: id.String(),
	}

	// Build payload with required fields
	payload := map[string]interface{}{
		"userId":    user.ID,
		"userEmail": user.Email,
		"expires":   expires,
	}

	// Add adminScopes if present
	if user.AdminScopes != nil {
		payload["adminScopes"] = user.AdminScopes
	}

	// Add any additional properties from extra
	if extra != nil {
		for key, value := range extra {
			payload[key] = value
		}
	}

	// Step 3: Base64URL encode header and payload
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT header: %w", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWT payload: %w", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Step 4: Sign
	toSign := headerB64 + "." + payloadB64
	signatureHmac := hmac.New(sha256.New, signingKey)
	signatureHmac.Write([]byte(toSign))
	signature := base64.RawURLEncoding.EncodeToString(signatureHmac.Sum(nil))

	jwt := toSign + "." + signature
	return jwt, nil
}

// apiRequest makes an HTTP request to the Vortex API
func (c *Client) apiRequest(method, path string, body interface{}, queryParams map[string]string) ([]byte, error) {
	// Build URL
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	if queryParams != nil {
		q := u.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		u.RawQuery = q.Encode()
	}

	// Prepare request body
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("User-Agent", userAgent)

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Vortex API request failed: %d %s", resp.StatusCode, resp.Status),
			Details:    string(responseBody),
		}
		return nil, apiErr
	}

	// Handle empty responses
	if len(responseBody) == 0 || string(responseBody) == "" {
		return []byte("{}"), nil
	}

	return responseBody, nil
}

// GetInvitationsByTarget retrieves invitations by target type and value
func (c *Client) GetInvitationsByTarget(targetType, targetValue string) ([]InvitationResult, error) {
	queryParams := map[string]string{
		"targetType":  targetType,
		"targetValue": targetValue,
	}

	responseBody, err := c.apiRequest("GET", "/api/v1/invitations", nil, queryParams)
	if err != nil {
		return nil, err
	}

	var response InvitationsResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Invitations, nil
}

// GetInvitation retrieves a specific invitation by ID
func (c *Client) GetInvitation(invitationID string) (*InvitationResult, error) {
	path := fmt.Sprintf("/api/v1/invitations/%s", invitationID)

	responseBody, err := c.apiRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var invitation InvitationResult
	if err := json.Unmarshal(responseBody, &invitation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &invitation, nil
}

// RevokeInvitation revokes an invitation
func (c *Client) RevokeInvitation(invitationID string) error {
	path := fmt.Sprintf("/api/v1/invitations/%s", invitationID)

	_, err := c.apiRequest("DELETE", path, nil, nil)
	return err
}

// AcceptInvitations accepts multiple invitations
func (c *Client) AcceptInvitations(invitationIDs []string, target InvitationTarget) (*InvitationResult, error) {
	requestBody := AcceptInvitationRequest{
		InvitationIDs: invitationIDs,
		Target:        target,
	}

	responseBody, err := c.apiRequest("POST", "/api/v1/invitations/accept", requestBody, nil)
	if err != nil {
		return nil, err
	}

	var result InvitationResult
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// DeleteInvitationsByGroup deletes all invitations for a specific group
func (c *Client) DeleteInvitationsByGroup(groupType, groupID string) error {
	path := fmt.Sprintf("/api/v1/invitations/by-group/%s/%s", groupType, groupID)

	_, err := c.apiRequest("DELETE", path, nil, nil)
	return err
}

// GetInvitationsByGroup retrieves invitations for a specific group
func (c *Client) GetInvitationsByGroup(groupType, groupID string) ([]InvitationResult, error) {
	path := fmt.Sprintf("/api/v1/invitations/by-group/%s/%s", groupType, groupID)

	responseBody, err := c.apiRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var response InvitationsResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Invitations, nil
}

// Reinvite sends a reinvitation for a specific invitation
func (c *Client) Reinvite(invitationID string) (*InvitationResult, error) {
	path := fmt.Sprintf("/api/v1/invitations/%s/reinvite", invitationID)

	responseBody, err := c.apiRequest("POST", path, nil, nil)
	if err != nil {
		return nil, err
	}

	var result InvitationResult
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}