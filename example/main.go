package main

import (
	"fmt"
	"log"
	"os"

	"github.com/TeamVortexSoftware/vortex-go-sdk"
)

func main() {
	// Initialize the client with API key from environment
	apiKey := os.Getenv("VORTEX_API_KEY")
	if apiKey == "" {
		apiKey = "demo-api-key"
	}

	client := vortex.NewClient(apiKey)

	// Example 1: Generate JWT - simple usage
	fmt.Println("=== JWT Generation Example ===")
	user := &vortex.User{
		ID:          "user-123",
		Email:       "admin@example.com",
		AdminScopes: []string{"autoJoin"},
	}

	jwt1, err := client.GenerateJWT(user, nil)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
	} else {
		fmt.Printf("Generated JWT: %s\n", jwt1)
	}

	// Example 1b: Generate JWT with additional properties
	fmt.Println("\n=== JWT Generation with Additional Properties ===")
	user2 := &vortex.User{
		ID:    "user-456",
		Email: "user@example.com",
	}

	extra := map[string]interface{}{
		"role":       "admin",
		"department": "Engineering",
	}

	jwt2, err := client.GenerateJWT(user2, extra)
	if err != nil {
		log.Printf("Failed to generate JWT with extra: %v", err)
	} else {
		fmt.Printf("Generated JWT with extra: %s\n", jwt2)
	}

	// Example 2: Get invitations by target
	fmt.Println("\n=== Get Invitations by Target Example ===")
	invitations, err := client.GetInvitationsByTarget("email", "user@example.com")
	if err != nil {
		if apiErr, ok := err.(*vortex.APIError); ok {
			fmt.Printf("API Error: %s (Status: %d)\n", apiErr.Message, apiErr.StatusCode)
			if apiErr.StatusCode == 404 {
				fmt.Println("This is expected with demo API key - showing empty results")
			}
		} else {
			fmt.Printf("Unexpected error: %s\n", err)
		}
	} else {
		fmt.Printf("Found %d invitations\n", len(invitations))
		for _, invitation := range invitations {
			fmt.Printf("- Invitation ID: %s, Status: %s\n", invitation.ID, invitation.Status)
		}
	}

	// Example 3: Get invitations by group
	fmt.Println("\n=== Get Invitations by Group Example ===")
	groupInvitations, err := client.GetInvitationsByGroup("team", "team-1")
	if err != nil {
		if apiErr, ok := err.(*vortex.APIError); ok {
			fmt.Printf("API Error: %s (Status: %d)\n", apiErr.Message, apiErr.StatusCode)
			if apiErr.StatusCode == 404 {
				fmt.Println("This is expected with demo API key - showing empty results")
			}
		} else {
			fmt.Printf("Unexpected error: %s\n", err)
		}
	} else {
		fmt.Printf("Found %d group invitations\n", len(groupInvitations))
		for _, invitation := range groupInvitations {
			fmt.Printf("- Invitation ID: %s, Status: %s\n", invitation.ID, invitation.Status)
		}
	}

	// Example 4: Accept invitations (will fail with demo API key, but shows usage)
	fmt.Println("\n=== Accept Invitations Example ===")
	target := vortex.InvitationTarget{
		Type:  "email",
		Value: "user@example.com",
	}

	result, err := client.AcceptInvitations([]string{"demo-invitation-id"}, target)
	if err != nil {
		if apiErr, ok := err.(*vortex.APIError); ok {
			fmt.Printf("API Error: %s (Status: %d)\n", apiErr.Message, apiErr.StatusCode)
			fmt.Println("This is expected with demo API key and fake invitation ID")
		} else {
			fmt.Printf("Unexpected error: %s\n", err)
		}
	} else {
		fmt.Printf("Accepted invitation: %s\n", result.ID)
	}

	fmt.Println("\n=== Example Complete ===")
	fmt.Println("To use with real data, set VORTEX_API_KEY environment variable")
}