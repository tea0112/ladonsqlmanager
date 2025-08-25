package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ladonsqlmanager"
	"github.com/ory/ladon"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🧪 Ladon Quick Test - Policy Scenarios")
	fmt.Println("======================================")

	// Connect to database
	dbString := os.Getenv("DB_STRING")
	if dbString == "" {
		log.Fatal("DB_STRING environment variable not set")
	}

	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize manager
	manager := ladonsqlmanager.New(db, "postgres")
	if err := manager.Init(); err != nil {
		log.Fatal("Failed to initialize manager:", err)
	}

	warden := &ladon.Ladon{Manager: manager}
	ctx := context.Background()

	// Test 1: Basic Policy Creation and Testing
	fmt.Println("\n🔐 Test 1: Basic Policy Creation")
	fmt.Println("--------------------------------")

	// Create a simple policy
	policy := &ladon.DefaultPolicy{
		ID:          "test-basic",
		Description: "Basic test policy",
		Effect:      ladon.AllowAccess,
		Subjects:    []string{"user"},
		Resources:   []string{"file:user:.*"},
		Actions:     []string{"read"},
	}

	if err := manager.Create(ctx, policy); err != nil {
		fmt.Printf("❌ Failed to create policy: %v\n", err)
	} else {
		fmt.Printf("✅ Created policy: %s\n", policy.ID)
	}

	// Test the policy
	request := &ladon.Request{
		Subject:  "user",
		Resource: "file:user:document.txt",
		Action:   "read",
	}

	if err := warden.IsAllowed(ctx, request); err != nil {
		fmt.Printf("❌ Access DENIED: %v\n", err)
	} else {
		fmt.Println("✅ Access ALLOWED for user reading own file")
	}

	// Test 2: Role-Based Access Control
	fmt.Println("\n👥 Test 2: Role-Based Access Control")
	fmt.Println("-----------------------------------")

	// Create admin policy
	adminPolicy := &ladon.DefaultPolicy{
		ID:          "admin-full",
		Description: "Admin has full access",
		Effect:      ladon.AllowAccess,
		Subjects:    []string{"admin"},
		Resources:   []string{".*"},
		Actions:     []string{".*"},
	}

	if err := manager.Create(ctx, adminPolicy); err != nil {
		fmt.Printf("❌ Failed to create admin policy: %v\n", err)
	} else {
		fmt.Printf("✅ Created admin policy: %s\n", adminPolicy.ID)
	}

	// Test admin access
	adminRequest := &ladon.Request{
		Subject:  "admin",
		Resource: "system:config",
		Action:   "delete",
	}

	if err := warden.IsAllowed(ctx, adminRequest); err != nil {
		fmt.Printf("❌ Admin access DENIED: %v\n", err)
	} else {
		fmt.Println("✅ Admin access ALLOWED to system config")
	}

	// Test user access to system resource
	userSystemRequest := &ladon.Request{
		Subject:  "user",
		Resource: "system:config",
		Action:   "delete",
	}

	if err := warden.IsAllowed(ctx, userSystemRequest); err != nil {
		fmt.Println("✅ User access correctly DENIED to system config")
	} else {
		fmt.Println("❌ User access incorrectly ALLOWED to system config")
	}

	// Test 3: Resource Hierarchy
	fmt.Println("\n🏗️ Test 3: Resource Hierarchy")
	fmt.Println("-----------------------------")

	// Create department-based policy
	deptPolicy := &ladon.DefaultPolicy{
		ID:          "dept-engineering",
		Description: "Engineering department access",
		Effect:      ladon.AllowAccess,
		Subjects:    []string{"dept:engineering"},
		Resources:   []string{"dept:engineering:.*"},
		Actions:     []string{"read", "write"},
	}

	if err := manager.Create(ctx, deptPolicy); err != nil {
		fmt.Printf("❌ Failed to create dept policy: %v\n", err)
	} else {
		fmt.Printf("✅ Created dept policy: %s\n", deptPolicy.ID)
	}

	// Test engineering access
	engRequest := &ladon.Request{
		Subject:  "dept:engineering",
		Resource: "dept:engineering:code",
		Action:   "write",
	}

	if err := warden.IsAllowed(ctx, engRequest); err != nil {
		fmt.Printf("❌ Engineering access DENIED: %v\n", err)
	} else {
		fmt.Println("✅ Engineering access ALLOWED to their code")
	}

	// Test cross-department access
	crossDeptRequest := &ladon.Request{
		Subject:  "dept:marketing",
		Resource: "dept:engineering:code",
		Action:   "read",
	}

	if err := warden.IsAllowed(ctx, crossDeptRequest); err != nil {
		fmt.Println("✅ Cross-department access correctly DENIED")
	} else {
		fmt.Println("❌ Cross-department access incorrectly ALLOWED")
	}

	// Test 4: Policy Listing and Details
	fmt.Println("\n📋 Test 4: Policy Management")
	fmt.Println("----------------------------")

	// List all policies
	policies, err := manager.GetAll(ctx, 100, 0)
	if err != nil {
		fmt.Printf("❌ Failed to get policies: %v\n", err)
	} else {
		fmt.Printf("📋 Found %d policies:\n", len(policies))
		for i, policy := range policies {
			fmt.Printf("  %d. %s (%s)\n", i+1, policy.GetID(), policy.GetDescription())
		}
	}

	// Test 5: Complex Authorization Scenarios
	fmt.Println("\n🎭 Test 5: Complex Authorization")
	fmt.Println("--------------------------------")

	// Create a policy with multiple subjects and actions
	complexPolicy := &ladon.DefaultPolicy{
		ID:          "complex-access",
		Description: "Complex access policy for multiple roles",
		Effect:      ladon.AllowAccess,
		Subjects:    []string{"user", "manager"},
		Resources:   []string{"api:user:.*", "api:public:.*"},
		Actions:     []string{"GET", "POST"},
	}

	if err := manager.Create(ctx, complexPolicy); err != nil {
		fmt.Printf("❌ Failed to create complex policy: %v\n", err)
	} else {
		fmt.Printf("✅ Created complex policy: %s\n", complexPolicy.ID)
	}

	// Test various combinations
	testCases := []struct {
		subject  string
		resource string
		action   string
		expected bool
	}{
		{"user", "api:user:profile", "GET", true},
		{"manager", "api:public:info", "POST", true},
		{"guest", "api:user:profile", "GET", false},
		{"user", "api:admin:users", "DELETE", false},
	}

	for _, tc := range testCases {
		request := &ladon.Request{
			Subject:  tc.subject,
			Resource: tc.resource,
			Action:   tc.action,
		}

		err := warden.IsAllowed(ctx, request)
		allowed := err == nil

		if allowed == tc.expected {
			fmt.Printf("✅ %s -> %s:%s = %v (expected: %v)\n",
				tc.subject, tc.resource, tc.action, allowed, tc.expected)
		} else {
			fmt.Printf("❌ %s -> %s:%s = %v (expected: %v)\n",
				tc.subject, tc.resource, tc.action, allowed, tc.expected)
		}
	}

	fmt.Println("\n🎉 Quick test completed!")
	fmt.Println("Use the interactive CLI for more experimentation:")
	fmt.Println("  go build -o playground/policy_cli playground/policy_cli.go")
	fmt.Println("  ./playground/policy_cli")
}
