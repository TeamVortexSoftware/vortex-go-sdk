package vortex

// InvitationTarget represents the target of an invitation
type InvitationTarget struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// InvitationGroup represents a group associated with an invitation
type InvitationGroup struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
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

// JWTPayload represents the payload for JWT generation
type JWTPayload struct {
	UserID      string      `json:"userId"`
	Identifiers []Identifier `json:"identifiers"`
	Groups      []Group     `json:"groups"`
	Role        *string     `json:"role,omitempty"`
}

// Identifier represents a user identifier (email, sms, etc.)
type Identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Group represents a user group
type Group struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JWTHeader represents the JWT header
type JWTHeader struct {
	IAT int64  `json:"iat"`
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

// JWTClaims represents the JWT payload claims
type JWTClaims struct {
	UserID      string      `json:"userId"`
	Groups      []Group     `json:"groups"`
	Role        *string     `json:"role,omitempty"`
	Expires     int64       `json:"expires"`
	Identifiers []Identifier `json:"identifiers"`
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