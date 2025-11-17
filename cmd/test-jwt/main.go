package main

import (
	"fmt"
	"os"

	vortex "github.com/teamvortexsoftware/vortex-go-sdk"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run test_jwt_main.go <api-key>")
		os.Exit(1)
	}

	apiKey := os.Args[1]
	client := vortex.NewClient(apiKey)

	// Test with admin scope
	userWithAdmin := &vortex.User{
		ID:          "test-user-123",
		Email:       "test@example.com",
		AdminScopes: []string{"autoJoin"},
	}
	jwtWithAdmin, err := client.GenerateJWT(userWithAdmin, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JWT (with admin): %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("WITH_ADMIN:%s\n", jwtWithAdmin)

	// Test without admin scope
	userNoAdmin := &vortex.User{
		ID:    "test-user-123",
		Email: "test@example.com",
	}
	jwtNoAdmin, err := client.GenerateJWT(userNoAdmin, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JWT (simple no admin): %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("NO_ADMIN:%s\n", jwtNoAdmin)

	// Test with extra properties
	extra := map[string]interface{}{
		"role":       "admin",
		"department": "Engineering",
	}
	jwtExtra, err := client.GenerateJWT(userWithAdmin, extra)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JWT (with extra): %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("WITH_EXTRA:%s\n", jwtExtra)
}
