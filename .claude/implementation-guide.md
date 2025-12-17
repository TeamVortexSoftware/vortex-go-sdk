# Vortex Go SDK Integration Guide

This guide provides step-by-step instructions for integrating Vortex into a Go application using the Go SDK.

## SDK Information

- **Package**: `github.com/TeamVortexSoftware/vortex-suite/packages/vortex-go-sdk`
- **Requires**: Go 1.18+
- **Dependencies**: `github.com/google/uuid v1.6.0`
- **Type**: Backend SDK with HTTP client

## Expected Input Context

This guide expects to receive the following context from the orchestrator:

### Integration Contract
```yaml
Integration Contract:
  API Endpoints:
    Prefix: /api/v1/vortex
    JWT: POST {prefix}/jwt
    Get Invitations: GET {prefix}/invitations
    Get Invitation: GET {prefix}/invitations/:id
    Accept Invitations: POST {prefix}/invitations/accept
  Scope:
    Entity: "workspace"
    Type: "workspace"
    ID Field: "workspace.id"
  File Paths:
    Backend:
      Vortex Handler: internal/handlers/vortex.go (or similar)
      Main App: cmd/server/main.go or main.go
      Routes: internal/routes/routes.go (or similar)
  Authentication:
    Pattern: "JWT Bearer token" (or session-based, etc.)
    User Extraction: Custom middleware/context
  Database:
    Library: "GORM" | "sqlx" | "database/sql" | "Ent" | "custom"
    User Model: users table/model
    Membership Model: workspace_members table/model (or equivalent)
```

### Discovery Data
- Backend technology stack (Go version, web framework)
- Web framework (Gin, Echo, Chi, Gorilla Mux, net/http, Fiber, etc.)
- Database library
- Authentication middleware in use
- Existing routing structure
- Environment variable management approach

## Implementation Overview

The Go SDK provides a client for JWT generation and API calls to Vortex. You'll need to:

1. Install the SDK
2. Create HTTP handlers that use the Vortex client
3. Implement custom logic for accepting invitations (database integration)
4. Register routes in your web framework

Unlike JavaScript SDKs, the Go SDK does NOT provide pre-built route handlers. You implement your own handlers using the Vortex client.

## Critical Go SDK Specifics

### Key Patterns
- **Client-Based**: Create a `vortex.Client` instance and call methods
- **Error Handling**: All methods return `error` - always check errors
- **Database Integration Required**: You must implement accept invitations logic
- **Framework Agnostic**: Works with any Go web framework (Gin, Echo, Chi, etc.)
- **JWT Generation**: `GenerateJWT(user, extra)` method with optional extra properties
- **API Methods**: `GetInvitationsByTarget`, `AcceptInvitations`, `GetInvitation`, etc.

### Basic Client Usage
```go
import "github.com/TeamVortexSoftware/vortex-suite/packages/vortex-go-sdk"

// Create client
client := vortex.NewClient(os.Getenv("VORTEX_API_KEY"))

// Generate JWT
user := &vortex.User{
    ID:          "user-123",
    Email:       "user@example.com",
    AdminScopes: []string{"autojoin"},
}

jwt, err := client.GenerateJWT(user, nil)
if err != nil {
    log.Fatal(err)
}
```

## Step-by-Step Implementation

### Step 1: Install SDK

```bash
go get github.com/TeamVortexSoftware/vortex-suite/packages/vortex-go-sdk
```

### Step 2: Set Up Environment Variables

Add to your `.env` file or environment:

```bash
VORTEX_API_KEY=VRTX.your-api-key-here.secret
```

**IMPORTANT**: Never commit your API key to version control.

### Step 3: Create Vortex Handler

Create a handler file (e.g., `internal/handlers/vortex.go`):

```go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	vortex "github.com/TeamVortexSoftware/vortex-suite/packages/vortex-go-sdk"
)

type VortexHandler struct {
	client *vortex.Client
	db     *YourDatabaseClient // Your database client (GORM, sqlx, etc.)
}

func NewVortexHandler(db *YourDatabaseClient) *VortexHandler {
	return &VortexHandler{
		client: vortex.NewClient(os.Getenv("VORTEX_API_KEY")),
		db:     db,
	}
}

// Helper function to send JSON error responses
func sendErrorJSON(w http.ResponseWriter, message, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
		"code":  code,
	})
}

// GetJWT generates a JWT for the authenticated user
func (h *VortexHandler) GetJWT(w http.ResponseWriter, r *http.Request) {
	// 0. Validate HTTP method
	if r.Method != http.MethodPost {
		sendErrorJSON(w, "Method not allowed", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	// 1. Extract authenticated user from context/session
	user := getUserFromContext(r.Context()) // Your auth function
	if user == nil {
		sendErrorJSON(w, "Unauthorized", "UNAUTHORIZED", http.StatusUnauthorized)
		return
	}

	// 2. Determine admin scopes
	var adminScopes []string
	if user.IsAdmin {
		adminScopes = []string{"autojoin"}
	}

	// 3. Generate JWT
	vortexUser := &vortex.User{
		ID:          user.ID,
		Email:       user.Email,
		AdminScopes: adminScopes,
	}

	jwt, err := h.client.GenerateJWT(vortexUser, nil)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		sendErrorJSON(w, "Failed to generate JWT", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	// 4. Return JWT
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"jwt": jwt})
}

// GetInvitationsByTarget retrieves invitations by target
func (h *VortexHandler) GetInvitationsByTarget(w http.ResponseWriter, r *http.Request) {
	// 1. Check authentication
	user := getUserFromContext(r.Context())
	if user == nil {
		sendErrorJSON(w, "Unauthorized", "UNAUTHORIZED", http.StatusUnauthorized)
		return
	}

	// 2. Get query parameters
	targetType := r.URL.Query().Get("targetType")
	targetValue := r.URL.Query().Get("targetValue")

	if targetType == "" || targetValue == "" {
		sendErrorJSON(w, "Missing targetType or targetValue", "INVALID_REQUEST", http.StatusBadRequest)
		return
	}

	// 3. Get invitations
	invitations, err := h.client.GetInvitationsByTarget(targetType, targetValue)
	if err != nil {
		log.Printf("Failed to get invitations: %v", err)
		sendErrorJSON(w, "Failed to get invitations", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	// 4. Return invitations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invitations": invitations,
	})
}

// GetInvitation retrieves a specific invitation by ID
func (h *VortexHandler) GetInvitation(w http.ResponseWriter, r *http.Request) {
	// 1. Check authentication
	user := getUserFromContext(r.Context())
	if user == nil {
		sendErrorJSON(w, "Unauthorized", "UNAUTHORIZED", http.StatusUnauthorized)
		return
	}

	// 2. Get invitation ID from URL path
	// This depends on your router (Chi, Gorilla, Gin, etc.)
	invitationID := getPathParam(r, "invitationId") // Your router's path param function

	// 3. Get invitation
	invitation, err := h.client.GetInvitation(invitationID)
	if err != nil {
		log.Printf("Failed to get invitation: %v", err)
		sendErrorJSON(w, "Failed to get invitation", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	// 4. Return invitation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invitation)
}

// AcceptInvitations accepts multiple invitations (CRITICAL - Custom Logic Required)
func (h *VortexHandler) AcceptInvitations(w http.ResponseWriter, r *http.Request) {
	// 1. Check authentication
	user := getUserFromContext(r.Context())
	if user == nil {
		sendErrorJSON(w, "Unauthorized", "UNAUTHORIZED", http.StatusUnauthorized)
		return
	}

	// 2. Parse request body
	var req struct {
		InvitationIDs []string             `json:"invitationIds"`
		Target        vortex.InvitationTarget `json:"target"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorJSON(w, "Invalid request body", "INVALID_REQUEST", http.StatusBadRequest)
		return
	}

	// 3. Accept invitations via Vortex API
	result, err := h.client.AcceptInvitations(req.InvitationIDs, req.Target)
	if err != nil {
		log.Printf("Failed to accept invitations: %v", err)
		sendErrorJSON(w, "Failed to accept invitations", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	// 4. Add user to your database for each group
	// THIS IS CRITICAL - ADJUST BASED ON YOUR DATABASE LIBRARY

	// Example with GORM:
	for _, group := range result.Groups {
		member := WorkspaceMember{
			UserID:      user.ID,
			WorkspaceID: group.GroupID, // Customer's group ID
			Role:        "member",
			JoinedAt:    time.Now(),
		}
		if err := h.db.Create(&member).Error; err != nil {
			log.Printf("Failed to create workspace member: %v", err)
			sendErrorJSON(w, "Failed to add user to workspace", "INTERNAL_ERROR", http.StatusInternalServerError)
			return
		}
	}

	// Example with sqlx:
	// for _, group := range result.Groups {
	//     _, err := h.db.Exec(
	//         "INSERT INTO workspace_members (user_id, workspace_id, role, joined_at) VALUES ($1, $2, $3, $4)",
	//         user.ID, group.GroupID, "member", time.Now(),
	//     )
	//     if err != nil {
	//         log.Printf("Failed to insert workspace member: %v", err)
	//         sendErrorJSON(w, "Failed to add user to workspace", "INTERNAL_ERROR", http.StatusInternalServerError)
	//         return
	//     }
	// }

	// Example with database/sql:
	// for _, group := range result.Groups {
	//     _, err := h.db.Exec(
	//         "INSERT INTO workspace_members (user_id, workspace_id, role, joined_at) VALUES (?, ?, ?, ?)",
	//         user.ID, group.GroupID, "member", time.Now(),
	//     )
	//     if err != nil {
	//         log.Printf("Failed to insert workspace member: %v", err)
	//         sendErrorJSON(w, "Failed to add user to workspace", "INTERNAL_ERROR", http.StatusInternalServerError)
	//         return
	//     }
	// }

	// 5. Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RevokeInvitation revokes an invitation
func (h *VortexHandler) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	// 1. Check authentication
	user := getUserFromContext(r.Context())
	if user == nil {
		sendErrorJSON(w, "Unauthorized", "UNAUTHORIZED", http.StatusUnauthorized)
		return
	}

	// 2. Get invitation ID
	invitationID := getPathParam(r, "invitationId")

	// 3. Revoke invitation
	if err := h.client.RevokeInvitation(invitationID); err != nil {
		log.Printf("Failed to revoke invitation: %v", err)
		sendErrorJSON(w, "Failed to revoke invitation", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	// 4. Return success
	w.WriteHeader(http.StatusNoContent)
}

// ReinviteUser resends an invitation
func (h *VortexHandler) ReinviteUser(w http.ResponseWriter, r *http.Request) {
	// 1. Check authentication
	user := getUserFromContext(r.Context())
	if user == nil {
		sendErrorJSON(w, "Unauthorized", "UNAUTHORIZED", http.StatusUnauthorized)
		return
	}

	// 2. Get invitation ID
	invitationID := getPathParam(r, "invitationId")

	// 3. Reinvite
	result, err := h.client.Reinvite(invitationID)
	if err != nil {
		log.Printf("Failed to reinvite: %v", err)
		sendErrorJSON(w, "Failed to reinvite", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	// 4. Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetInvitationsByGroup retrieves invitations for a group
func (h *VortexHandler) GetInvitationsByGroup(w http.ResponseWriter, r *http.Request) {
	// 1. Check authentication
	user := getUserFromContext(r.Context())
	if user == nil {
		sendErrorJSON(w, "Unauthorized", "UNAUTHORIZED", http.StatusUnauthorized)
		return
	}

	// 2. Get path parameters
	groupType := getPathParam(r, "groupType")
	groupID := getPathParam(r, "groupId")

	// 3. Get invitations
	invitations, err := h.client.GetInvitationsByGroup(groupType, groupID)
	if err != nil {
		log.Printf("Failed to get invitations by group: %v", err)
		sendErrorJSON(w, "Failed to get invitations", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	// 4. Return invitations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invitations": invitations,
	})
}

// DeleteInvitationsByGroup deletes all invitations for a group
func (h *VortexHandler) DeleteInvitationsByGroup(w http.ResponseWriter, r *http.Request) {
	// 1. Check authentication
	user := getUserFromContext(r.Context())
	if user == nil {
		sendErrorJSON(w, "Unauthorized", "UNAUTHORIZED", http.StatusUnauthorized)
		return
	}

	// 2. Get path parameters
	groupType := getPathParam(r, "groupType")
	groupID := getPathParam(r, "groupId")

	// 3. Delete invitations
	if err := h.client.DeleteInvitationsByGroup(groupType, groupID); err != nil {
		log.Printf("Failed to delete invitations by group: %v", err)
		sendErrorJSON(w, "Failed to delete invitations", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	// 4. Return success
	w.WriteHeader(http.StatusNoContent)
}
```

### Step 4: Register Routes

The route registration depends on your web framework. Here are examples for common frameworks:

#### net/http (Standard Library)

```go
package main

import (
	"net/http"
	"your-app/internal/handlers"
)

func main() {
	// Create handler
	vortexHandler := handlers.NewVortexHandler(db)

	// Register routes
	http.HandleFunc("/api/v1/vortex/jwt", vortexHandler.GetJWT)
	http.HandleFunc("/api/v1/vortex/invitations", vortexHandler.GetInvitationsByTarget)
	http.HandleFunc("/api/v1/vortex/invitations/accept", vortexHandler.AcceptInvitations)
	// Note: net/http doesn't have great path param support - consider using a router

	http.ListenAndServe(":8080", nil)
}
```

#### Chi Router (Recommended)

```go
package main

import (
	"net/http"
	"your-app/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Your authentication middleware
	r.Use(authMiddleware)

	// Create handler
	vortexHandler := handlers.NewVortexHandler(db)

	// Register Vortex routes
	r.Route("/api/v1/vortex", func(r chi.Router) {
		r.Post("/jwt", vortexHandler.GetJWT)
		r.Get("/invitations", vortexHandler.GetInvitationsByTarget)
		r.Post("/invitations/accept", vortexHandler.AcceptInvitations)
		r.Get("/invitations/{invitationId}", vortexHandler.GetInvitation)
		r.Delete("/invitations/{invitationId}", vortexHandler.RevokeInvitation)
		r.Post("/invitations/{invitationId}/reinvite", vortexHandler.ReinviteUser)
		r.Get("/invitations/by-group/{groupType}/{groupId}", vortexHandler.GetInvitationsByGroup)
		r.Delete("/invitations/by-group/{groupType}/{groupId}", vortexHandler.DeleteInvitationsByGroup)
	})

	http.ListenAndServe(":8080", r)
}

// Helper function for Chi router
func getPathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
```

#### Gin Framework

```go
package main

import (
	"your-app/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Your authentication middleware
	r.Use(authMiddleware())

	// Create handler
	vortexHandler := handlers.NewVortexHandler(db)

	// Register Vortex routes
	vortex := r.Group("/api/v1/vortex")
	{
		vortex.POST("/jwt", ginAdapter(vortexHandler.GetJWT))
		vortex.GET("/invitations", ginAdapter(vortexHandler.GetInvitationsByTarget))
		vortex.POST("/invitations/accept", ginAdapter(vortexHandler.AcceptInvitations))
		vortex.GET("/invitations/:invitationId", ginAdapter(vortexHandler.GetInvitation))
		vortex.DELETE("/invitations/:invitationId", ginAdapter(vortexHandler.RevokeInvitation))
		vortex.POST("/invitations/:invitationId/reinvite", ginAdapter(vortexHandler.ReinviteUser))
		vortex.GET("/invitations/by-group/:groupType/:groupId", ginAdapter(vortexHandler.GetInvitationsByGroup))
		vortex.DELETE("/invitations/by-group/:groupType/:groupId", ginAdapter(vortexHandler.DeleteInvitationsByGroup))
	}

	r.Run(":8080")
}

// Adapter to convert http.HandlerFunc to gin.HandlerFunc
func ginAdapter(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h(c.Writer, c.Request)
	}
}

// Helper for Gin
func getPathParam(r *http.Request, key string) string {
	// In Gin, you'd typically use c.Param(key) in the handler directly
	// Or store it in context
	return r.Context().Value(key).(string)
}
```

#### Echo Framework

```go
package main

import (
	"net/http"
	"your-app/internal/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Your authentication middleware
	e.Use(authMiddleware)

	// Create handler
	vortexHandler := handlers.NewVortexHandler(db)

	// Register Vortex routes
	vortex := e.Group("/api/v1/vortex")
	{
		vortex.POST("/jwt", echoAdapter(vortexHandler.GetJWT))
		vortex.GET("/invitations", echoAdapter(vortexHandler.GetInvitationsByTarget))
		vortex.POST("/invitations/accept", echoAdapter(vortexHandler.AcceptInvitations))
		vortex.GET("/invitations/:invitationId", echoAdapter(vortexHandler.GetInvitation))
		vortex.DELETE("/invitations/:invitationId", echoAdapter(vortexHandler.RevokeInvitation))
		vortex.POST("/invitations/:invitationId/reinvite", echoAdapter(vortexHandler.ReinviteUser))
		vortex.GET("/invitations/by-group/:groupType/:groupId", echoAdapter(vortexHandler.GetInvitationsByGroup))
		vortex.DELETE("/invitations/by-group/:groupType/:groupId", echoAdapter(vortexHandler.DeleteInvitationsByGroup))
	}

	e.Start(":8080")
}

// Adapter to convert http.HandlerFunc to echo.HandlerFunc
func echoAdapter(h http.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		h(c.Response(), c.Request())
		return nil
	}
}

// Helper for Echo
func getPathParam(r *http.Request, key string) string {
	// In Echo, you'd typically use c.Param(key) in the handler directly
	return r.Context().Value(key).(string)
}
```

### Step 5: Add CORS Configuration (If Needed)

If your frontend is on a different domain, add CORS middleware:

#### Chi Router

```go
import "github.com/go-chi/cors"

r.Use(cors.Handler(cors.Options{
	AllowedOrigins:   []string{os.Getenv("FRONTEND_URL")},
	AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
	AllowCredentials: true,
}))
```

#### Gin

```go
import "github.com/gin-contrib/cors"

r.Use(cors.New(cors.Config{
	AllowOrigins:     []string{os.Getenv("FRONTEND_URL")},
	AllowMethods:     []string{"GET", "POST", "DELETE"},
	AllowHeaders:     []string{"Content-Type", "Authorization"},
	AllowCredentials: true,
}))
```

## Build and Validation

### Build Your Application

```bash
go build -o server cmd/server/main.go
# or
go build ./...
```

### Test the Integration

Start your server and test each endpoint:

```bash
# Start server
./server

# Test JWT endpoint
curl -X POST http://localhost:8080/api/v1/vortex/jwt \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN"

# Test get invitations
curl -X GET "http://localhost:8080/api/v1/vortex/invitations?targetType=email&targetValue=user@example.com" \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN"

# Test accept invitations
curl -X POST http://localhost:8080/api/v1/vortex/invitations/accept \
  -H "Authorization: Bearer YOUR_AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "invitationIds": ["invitation-id-1"],
    "target": { "type": "email", "value": "user@example.com" }
  }'
```

### Validation Checklist

- [ ] SDK installed successfully (`go get`)
- [ ] Environment variable `VORTEX_API_KEY` is set
- [ ] Vortex handler created with all methods
- [ ] Routes registered in web framework
- [ ] JWT endpoint returns valid JWT
- [ ] Accept invitations endpoint adds users to database
- [ ] Authentication middleware protects all Vortex endpoints
- [ ] CORS is configured (if frontend on different domain)
- [ ] Code compiles without errors (`go build`)

## Implementation Report

After completing the integration, provide this summary:

```markdown
## Go SDK Integration Complete

### Files Modified/Created
- `internal/handlers/vortex.go` - Vortex handlers using Go SDK client
- `cmd/server/main.go` - Registered Vortex routes at /api/v1/vortex
- `.env` - Added VORTEX_API_KEY environment variable

### Endpoints Registered
- POST /api/v1/vortex/jwt - Generate JWT for authenticated user
- GET /api/v1/vortex/invitations - Get invitations by target
- GET /api/v1/vortex/invitations/:id - Get invitation by ID
- POST /api/v1/vortex/invitations/accept - Accept invitations (custom logic)
- DELETE /api/v1/vortex/invitations/:id - Revoke invitation
- POST /api/v1/vortex/invitations/:id/reinvite - Resend invitation
- GET /api/v1/vortex/invitations/by-group/:type/:id - Get invitations for group
- DELETE /api/v1/vortex/invitations/by-group/:type/:id - Delete invitations for group

### Database Integration
- Library: [GORM/sqlx/database/sql/etc.]
- Accept invitations adds users to: [workspace_members table]
- Group association field: [workspace_id/team_id/etc.]

### Web Framework
- Framework: [Chi/Gin/Echo/net/http/etc.]
- Authentication: [JWT/Session/Custom middleware]
- User extraction: [Context value/custom function]

### Next Steps for Frontend
The backend now exposes these endpoints for the frontend to consume:
1. Call POST /api/v1/vortex/jwt to get JWT for Vortex widget
2. Pass JWT to Vortex widget component
3. Widget will handle invitation sending
4. Accept invitations via POST /api/v1/vortex/invitations/accept
```

## Common Issues and Solutions

### Issue: "cannot find package"
**Solution**: Ensure the SDK is installed:
```bash
go get github.com/TeamVortexSoftware/vortex-suite/packages/vortex-go-sdk
go mod tidy
```

### Issue: "undefined: vortex.NewClient"
**Solution**: Make sure you're importing the correct package:
```go
import vortex "github.com/TeamVortexSoftware/vortex-suite/packages/vortex-go-sdk"
```

### Issue: "user is nil in handlers"
**Solution**: Ensure your authentication middleware populates the user in context before Vortex handlers run.

### Issue: "path parameters not working"
**Solution**: The way to extract path parameters depends on your router:
- Chi: `chi.URLParam(r, "invitationId")`
- Gin: `c.Param("invitationId")` (in Gin context)
- Echo: `c.Param("invitationId")` (in Echo context)
- Gorilla Mux: `mux.Vars(r)["invitationId"]`

### Issue: "CORS errors from frontend"
**Solution**: Add CORS middleware for your framework (see Step 5).

### Issue: "Accept invitations succeeds but user not added to database"
**Solution**: You must implement custom database logic in the `AcceptInvitations` handler (see Step 3).

### Issue: "JWT generation fails with invalid API key"
**Solution**: Ensure your API key is in the correct format: `VRTX.base64id.key`

## Best Practices

### 1. Environment Variables
Use environment variables for sensitive configuration:
```go
import "github.com/joho/godotenv"

func init() {
	godotenv.Load()
}

apiKey := os.Getenv("VORTEX_API_KEY")
if apiKey == "" {
	log.Fatal("VORTEX_API_KEY is required")
}
```

### 2. Error Handling
Always handle errors from Vortex client methods:
```go
jwt, err := client.GenerateJWT(user, nil)
if err != nil {
	if apiErr, ok := err.(*vortex.APIError); ok {
		log.Printf("API Error: %s (Status: %d)", apiErr.Message, apiErr.StatusCode)
	} else {
		log.Printf("Unexpected error: %v", err)
	}
	sendErrorJSON(w, "Internal server error", "INTERNAL_ERROR", http.StatusInternalServerError)
	return
}
```

### 3. Struct Validation
Validate request bodies before processing:
```go
if len(req.InvitationIDs) == 0 {
	sendErrorJSON(w, "invitationIds is required", "INVALID_REQUEST", http.StatusBadRequest)
	return
}

if req.Target.Type == "" || req.Target.Value == "" {
	sendErrorJSON(w, "target type and value are required", "INVALID_REQUEST", http.StatusBadRequest)
	return
}
```

### 4. Database Transactions
Use transactions for accept invitations:
```go
// GORM example
tx := h.db.Begin()
defer func() {
	if r := recover(); r != nil {
		tx.Rollback()
	}
}()

for _, group := range result.Groups {
	member := WorkspaceMember{UserID: user.ID, WorkspaceID: group.GroupID, Role: "member"}
	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback()
		sendErrorJSON(w, "Failed to add user", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}
}

tx.Commit()
```

### 5. Admin Scopes
Only grant autojoin to administrators:
```go
var adminScopes []string
if user.Role == "admin" {
	adminScopes = []string{"autojoin"}
}

vortexUser := &vortex.User{
	ID:          user.ID,
	Email:       user.Email,
	AdminScopes: adminScopes,
}
```

### 6. Logging
Add structured logging:
```go
import "log/slog"

slog.Info("Generating JWT", "userId", user.ID)
slog.Error("Failed to accept invitations", "error", err, "userId", user.ID)
```

### 7. Context Usage
Use context for request-scoped values:
```go
type contextKey string

const userContextKey contextKey = "user"

func getUserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}
```

## Additional Resources

- [Go SDK Documentation](https://docs.vortexsoftware.com/sdks/go)
- [Vortex API Reference](https://api.vortexsoftware.com/api)
- [Go Web Frameworks Comparison](https://github.com/mingrammer/go-web-framework-stars)
- [Integration Examples](https://github.com/teamvortexsoftware/vortex-examples)

## Support

For questions or issues:
- GitHub Issues: https://github.com/TeamVortexSoftware/vortex-suite/issues
- Email: support@vortexsoftware.com
- Documentation: https://docs.vortexsoftware.com
