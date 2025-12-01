package vortex

// User represents user data for JWT generation
type User struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	AdminScopes []string `json:"adminScopes,omitempty"`
}

// InvitationTarget represents the target of an invitation
type InvitationTarget struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// InvitationGroup represents a group associated with an invitation
// This matches the MemberGroups table structure from the API response
type InvitationGroup struct {
	ID        string `json:"id"`        // Vortex internal UUID
	AccountID string `json:"accountId"` // Vortex account ID
	GroupID   string `json:"groupId"`   // Customer's group ID (the ID they provided)
	Type      string `json:"type"`      // Group type (e.g., "workspace", "team")
	Name      string `json:"name"`      // Group name
	CreatedAt string `json:"createdAt"` // Timestamp when the group was created
}

// InvitationAcceptance represents an accepted invitation
type InvitationAcceptance struct {
	ID         string            `json:"id"`
	AccountID  string            `json:"accountId"`
	ProjectID  string            `json:"projectId"`
	AcceptedAt string            `json:"acceptedAt"`
	Target     InvitationTarget  `json:"target"`
}

// InvitationResult represents a complete invitation object
type InvitationResult struct {
	ID                       string                  `json:"id"`
	AccountID                string                  `json:"accountId"`
	ClickThroughs            int                     `json:"clickThroughs"`
	ConfigurationAttributes  map[string]interface{}  `json:"configurationAttributes"`
	Attributes               map[string]interface{}  `json:"attributes"`
	CreatedAt                string                  `json:"createdAt"`
	Deactivated              bool                    `json:"deactivated"`
	DeliveryCount            int                     `json:"deliveryCount"`
	DeliveryTypes            []string                `json:"deliveryTypes"`
	ForeignCreatorID         string                  `json:"foreignCreatorId"`
	InvitationType           string                  `json:"invitationType"`
	ModifiedAt               *string                 `json:"modifiedAt"`
	Status                   string                  `json:"status"`
	Target                   []InvitationTarget      `json:"target"`
	Views                    int                     `json:"views"`
	WidgetConfigurationID    string                  `json:"widgetConfigurationId"`
	ProjectID                string                  `json:"projectId"`
	Groups                   []InvitationGroup       `json:"groups"`
	Accepts                  []InvitationAcceptance  `json:"accepts"`
	Expired                  bool                    `json:"expired"`
	Expires                  *string                 `json:"expires,omitempty"`
}

// AcceptInvitationRequest represents the request body for accepting invitations
type AcceptInvitationRequest struct {
	InvitationIDs []string         `json:"invitationIds"`
	Target        InvitationTarget `json:"target"`
}

// InvitationsResponse represents the API response containing multiple invitations
type InvitationsResponse struct {
	Invitations []InvitationResult `json:"invitations"`
}

// JWTPayload represents the payload for JWT generation (legacy format)
// Deprecated: Use JWTPayloadSimple for new implementations
type JWTPayload struct {
	UserID      string      `json:"userId"`
	Identifiers []Identifier `json:"identifiers"`
	Groups      []Group     `json:"groups"`
	Role        *string     `json:"role,omitempty"`
}

// JWTPayloadSimple represents the simplified JWT payload (recommended)
type JWTPayloadSimple struct {
	UserID              string `json:"userId"`
	UserEmail           string `json:"userEmail"`
	UserIsAutoJoinAdmin *bool  `json:"userIsAutoJoinAdmin,omitempty"`
}

// Identifier represents a user identifier (email, sms, etc.)
type Identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Group represents a user group for JWT generation (input)
// For backward compatibility, supports both 'id' and 'groupId' fields
type Group struct {
	Type    string  `json:"type"`
	ID      *string `json:"id,omitempty"`      // Legacy field (deprecated, use GroupID)
	GroupID *string `json:"groupId,omitempty"` // Preferred: Customer's group ID
	Name    string  `json:"name"`
}

// JWTHeader represents the JWT header
type JWTHeader struct {
	IAT int64  `json:"iat"`
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

// JWTClaims represents the JWT payload claims
// Supports both new simplified format and legacy format
type JWTClaims struct {
	UserID              string       `json:"userId"`
	UserEmail           string       `json:"userEmail,omitempty"`
	UserIsAutoJoinAdmin *bool        `json:"userIsAutoJoinAdmin,omitempty"`
	Groups              []Group      `json:"groups,omitempty"`
	Role                *string      `json:"role,omitempty"`
	Expires             int64        `json:"expires"`
	Identifiers         []Identifier `json:"identifiers,omitempty"`
}

// APIError represents an error from the Vortex API
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}