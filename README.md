# Vortex Go SDK

A Go SDK for Vortex invitation management and JWT generation.

## Installation

```bash
go get github.com/TeamVortexSoftware/vortex-go-sdk
```

## Usage

### Basic Setup

```go
package main

import (
    "fmt"
    "log"

    "https://github.com/TeamVortexSoftware/vortex-go-sdk"
)

func main() {
    // Initialize the client
    client := vortex.NewClient("your-api-key")
}
```

### JWT Generation

```go
// Simple usage
user := &vortex.User{
    ID:          "user-123",
    Email:       "user@example.com",
    AdminScopes: []string{"autoJoin"},
}

jwt, err := client.GenerateJWT(user, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("JWT: %s\n", jwt)

// With additional properties
extra := map[string]interface{}{
    "role":       "admin",
    "department": "Engineering",
}

jwt, err = client.GenerateJWT(user, extra)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("JWT with extra: %s\n", jwt)
```

### Invitation Management

#### Get Invitations by Target

```go
// Get invitations for a specific target
invitations, err := client.GetInvitationsByTarget("email", "user@example.com")
if err != nil {
    log.Fatal(err)
}

for _, invitation := range invitations {
    fmt.Printf("Invitation ID: %s, Status: %s\n", invitation.ID, invitation.Status)
}
```

#### Accept Invitations

```go
// Accept multiple invitations
target := vortex.InvitationTarget{
    Type:  "email",
    Value: "user@example.com",
}

result, err := client.AcceptInvitations([]string{"inv1", "inv2"}, target)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Accepted invitation: %s\n", result.ID)
```

#### Get Specific Invitation

```go
// Get a specific invitation by ID
invitation, err := client.GetInvitation("invitation-id")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Invitation: %s\n", invitation.ID)
```

#### Revoke Invitation

```go
// Revoke an invitation
err := client.RevokeInvitation("invitation-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Invitation revoked successfully")
```

### Group Operations

#### Get Invitations by Group

```go
// Get invitations for a specific group
invitations, err := client.GetInvitationsByGroup("organization", "org123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d invitations\n", len(invitations))
```

#### Delete Invitations by Group

```go
// Delete all invitations for a group
err := client.DeleteInvitationsByGroup("organization", "org123")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Group invitations deleted successfully")
```

#### Reinvite

```go
// Send a reinvitation
invitation, err := client.Reinvite("invitation-id")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Reinvited: %s\n", invitation.ID)
```

## Error Handling

The SDK returns custom error types that provide detailed information about API failures:

```go
invitations, err := client.GetInvitationsByTarget("email", "user@example.com")
if err != nil {
    if apiErr, ok := err.(*vortex.APIError); ok {
        fmt.Printf("API Error: %s (Status: %d)\n", apiErr.Message, apiErr.StatusCode)
        fmt.Printf("Details: %s\n", apiErr.Details)
    } else {
        fmt.Printf("Unexpected error: %s\n", err)
    }
    return
}
```

## Environment Variables

- `VORTEX_API_BASE_URL` - Base URL for Vortex API (default: https://api.vortexsoftware.com)

## API Compatibility

This Go SDK provides identical functionality to the Node.js SDK:

- Same JWT generation algorithm with HMAC-SHA256
- Same API endpoints and request/response formats
- Same error handling patterns
- Compatible with Express, Fastify, Next.js, and Python SDKs

## Data Types

### Core Types

```go
// InvitationTarget represents the target of an invitation
type InvitationTarget struct {
    Type  string `json:"type"`  // "email", "sms", "username", "phoneNumber"
    Value string `json:"value"`
}

// InvitationGroup represents a group associated with an invitation
type InvitationGroup struct {
    ID        string `json:"id"`        // Vortex internal UUID
    AccountID string `json:"accountId"` // Vortex account ID
    GroupID   string `json:"groupId"`   // Customer's group ID (the ID they provided)
    Type      string `json:"type"`      // Group type (e.g., "workspace", "team")
    Name      string `json:"name"`      // Group name
    CreatedAt string `json:"createdAt"` // Timestamp when the group was created
}

// InvitationResult represents a complete invitation object
type InvitationResult struct {
    ID                    string                 `json:"id"`
    AccountID             string                 `json:"accountId"`
    ClickThroughs         int                    `json:"clickThroughs"`
    ConfigurationAttributes map[string]interface{} `json:"configurationAttributes"`
    Attributes            map[string]interface{} `json:"attributes"`
    CreatedAt             string                 `json:"createdAt"`
    Deactivated           bool                   `json:"deactivated"`
    DeliveryCount         int                    `json:"deliveryCount"`
    DeliveryTypes         []string               `json:"deliveryTypes"`
    ForeignCreatorID      string                 `json:"foreignCreatorId"`
    InvitationType        string                 `json:"invitationType"`
    ModifiedAt            *string                `json:"modifiedAt"`
    Status                string                 `json:"status"`
    Target                []InvitationTarget     `json:"target"`
    Views                 int                    `json:"views"`
    WidgetConfigurationID string                 `json:"widgetConfigurationId"`
    ProjectID             string                 `json:"projectId"`
    Groups                []InvitationGroup      `json:"groups"`
    Accepts               []InvitationAcceptance `json:"accepts"`
}
```

### JWT Types

```go
// User represents user data for JWT generation
type User struct {
    ID          string   `json:"id"`
    Email       string   `json:"email"`
    AdminScopes []string `json:"adminScopes,omitempty"`
}
```

The `AdminScopes` field is optional. If provided, the full array will be included in the JWT payload as `adminScopes`.

## Development

### Building

```bash
go build ./...
```

### Running Tests

```bash
go test ./...
```

### Module Dependencies

- Go 1.18+
- github.com/google/uuid v1.6.0

## License

MIT