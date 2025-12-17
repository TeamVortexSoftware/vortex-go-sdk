//go:build !integration
// +build !integration

package vortex

import (
	"fmt"
	"os"
)

// TestJWTVerify is a simple program to test JWT generation
// Run with: go run jwt_verify_test.go <api-key>
func TestJWTVerify() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go test -run=TestJWTVerify <api-key>")
		os.Exit(1)
	}

	apiKey := os.Args[1]
	client := NewClient(apiKey)

	// Test simplified format with admin flag
	boolTrue := true
	jwtSimpleWithAdmin, err := client.GenerateJWTSimple("test-user-123", "test@example.com", &boolTrue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JWT (simple with admin): %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("SIMPLE_WITH_ADMIN:%s\n", jwtSimpleWithAdmin)

	// Test simplified format without admin flag
	jwtSimpleNoAdmin, err := client.GenerateJWTSimple("test-user-123", "test@example.com", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JWT (simple no admin): %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("SIMPLE_NO_ADMIN:%s\n", jwtSimpleNoAdmin)

	// Test legacy format
	role := "admin"
	jwtLegacy, err := client.GenerateJWT(JWTPayload{
		UserID: "test-user-123",
		Identifiers: []Identifier{
			{Type: "email", Value: "test@example.com"},
		},
		Groups: []Group{
			{Type: "team", GroupID: "team-123", Name: "Test Team"},
		},
		Role: &role,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JWT (legacy): %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("LEGACY:%s\n", jwtLegacy)
}
