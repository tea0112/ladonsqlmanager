package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ladonsqlmanager"
	"github.com/ory/ladon"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PolicyCLI struct {
	manager *ladonsqlmanager.SQLManager
	warden  *ladon.Ladon
	reader  *bufio.Reader
}

func NewPolicyCLI(db *gorm.DB) *PolicyCLI {
	manager := ladonsqlmanager.New(db, "postgres")
	if err := manager.Init(); err != nil {
		log.Fatal("Failed to initialize manager:", err)
	}

	return &PolicyCLI{
		manager: manager,
		warden:  &ladon.Ladon{Manager: manager},
		reader:  bufio.NewReader(os.Stdin),
	}
}

func (cli *PolicyCLI) Run() {
	fmt.Println("🎮 Welcome to Ladon Policy Playground!")
	fmt.Println("=====================================")

	for {
		cli.showMenu()
		choice := cli.readChoice()

		switch choice {
		case 1:
			cli.createPolicy()
		case 2:
			cli.listPolicies()
		case 3:
			cli.testAuthorization()
		case 4:
			cli.createSamplePolicies()
		case 5:
			cli.deletePolicy()
		case 6:
			cli.showPolicyDetails()
		case 0:
			fmt.Println("👋 Goodbye!")
			return
		default:
			fmt.Println("❌ Invalid choice. Please try again.")
		}
		fmt.Println()
	}
}

func (cli *PolicyCLI) showMenu() {
	fmt.Println("\n📋 Available Actions:")
	fmt.Println("1. Create a new policy")
	fmt.Println("2. List all policies")
	fmt.Println("3. Test authorization")
	fmt.Println("4. Create sample policies")
	fmt.Println("5. Delete a policy")
	fmt.Println("6. Show policy details")
	fmt.Println("0. Exit")
	fmt.Print("\nEnter your choice: ")
}

func (cli *PolicyCLI) readChoice() int {
	input, _ := cli.reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return -1
	}
	return choice
}

func (cli *PolicyCLI) createPolicy() {
	fmt.Println("\n🔐 Creating New Policy")
	fmt.Println("=====================")

	policy := &ladon.DefaultPolicy{}

	fmt.Print("Policy ID: ")
	policy.ID = cli.readString()

	fmt.Print("Description: ")
	policy.Description = cli.readString()

	fmt.Print("Effect (allow/deny): ")
	effect := cli.readString()
	if effect == "allow" {
		policy.Effect = ladon.AllowAccess
	} else {
		policy.Effect = ladon.DenyAccess
	}

	fmt.Print("Subjects (comma-separated): ")
	subjects := cli.readString()
	policy.Subjects = strings.Split(subjects, ",")

	fmt.Print("Resources (comma-separated): ")
	resources := cli.readString()
	policy.Resources = strings.Split(resources, ",")

	fmt.Print("Actions (comma-separated): ")
	actions := cli.readString()
	policy.Actions = strings.Split(actions, ",")

	ctx := context.Background()
	if err := cli.manager.Create(ctx, policy); err != nil {
		fmt.Printf("❌ Failed to create policy: %v\n", err)
	} else {
		fmt.Printf("✅ Policy '%s' created successfully!\n", policy.ID)
	}
}

func (cli *PolicyCLI) listPolicies() {
	fmt.Println("\n📋 All Policies")
	fmt.Println("===============")

	ctx := context.Background()
	policies, err := cli.manager.GetAll(ctx, 100, 0)
	if err != nil {
		fmt.Printf("❌ Failed to get policies: %v\n", err)
		return
	}

	if len(policies) == 0 {
		fmt.Println("No policies found.")
		return
	}

	for i, policy := range policies {
		fmt.Printf("%d. %s (%s)\n", i+1, policy.GetID(), policy.GetDescription())
		fmt.Printf("   Effect: %s\n", policy.GetEffect())
		fmt.Printf("   Subjects: %v\n", policy.GetSubjects())
		fmt.Printf("   Resources: %v\n", policy.GetResources())
		fmt.Printf("   Actions: %v\n", policy.GetActions())
		fmt.Println()
	}
}

func (cli *PolicyCLI) testAuthorization() {
	fmt.Println("\n🧪 Test Authorization")
	fmt.Println("====================")

	request := &ladon.Request{}

	fmt.Print("Subject: ")
	request.Subject = cli.readString()

	fmt.Print("Resource: ")
	request.Resource = cli.readString()

	fmt.Print("Action: ")
	request.Action = cli.readString()

	ctx := context.Background()
	err := cli.warden.IsAllowed(ctx, request)

	if err != nil {
		fmt.Printf("❌ Access DENIED: %v\n", err)
	} else {
		fmt.Println("✅ Access ALLOWED!")
	}

	// Show matching policies
	policies, err := cli.warden.Manager.FindRequestCandidates(ctx, request)
	if err != nil {
		fmt.Printf("❌ Failed to find policies: %v\n", err)
		return
	}

	fmt.Printf("📋 Found %d matching policies:\n", len(policies))
	for i, policy := range policies {
		fmt.Printf("  %d. %s (%s)\n", i+1, policy.GetID(), policy.GetDescription())
	}
}

func (cli *PolicyCLI) createSamplePolicies() {
	fmt.Println("\n🎯 Creating Sample Policies")
	fmt.Println("==========================")

	samples := []*ladon.DefaultPolicy{
		{
			ID:          "admin-full-access",
			Description: "Admin has full access to everything",
			Effect:      ladon.AllowAccess,
			Subjects:    []string{"admin"},
			Resources:   []string{".*"},
			Actions:     []string{".*"},
		},
		{
			ID:          "user-read-own-files",
			Description: "Users can read their own files",
			Effect:      ladon.AllowAccess,
			Subjects:    []string{"user"},
			Resources:   []string{"file:user:<.*>"},
			Actions:     []string{"read"},
		},
		{
			ID:          "guest-public-read",
			Description: "Guests can read public resources",
			Effect:      ladon.AllowAccess,
			Subjects:    []string{"guest"},
			Resources:   []string{"public:.*"},
			Actions:     []string{"read"},
		},
		{
			ID:          "deny-guest-write",
			Description: "Guests cannot write to any resource",
			Effect:      ladon.DenyAccess,
			Subjects:    []string{"guest"},
			Resources:   []string{".*"},
			Actions:     []string{"write", "delete", "create"},
		},
	}

	ctx := context.Background()
	created := 0

	for _, policy := range samples {
		if err := cli.manager.Create(ctx, policy); err != nil {
			fmt.Printf("❌ Failed to create %s: %v\n", policy.ID, err)
		} else {
			fmt.Printf("✅ Created %s\n", policy.ID)
			created++
		}
	}

	fmt.Printf("\n🎉 Created %d sample policies!\n", created)
}

func (cli *PolicyCLI) deletePolicy() {
	fmt.Println("\n🗑️ Delete Policy")
	fmt.Println("================")

	fmt.Print("Enter policy ID to delete: ")
	policyID := cli.readString()

	if policyID == "" {
		fmt.Println("❌ Policy ID cannot be empty")
		return
	}

	ctx := context.Background()
	if err := cli.manager.Delete(ctx, policyID); err != nil {
		fmt.Printf("❌ Failed to delete policy: %v\n", err)
	} else {
		fmt.Printf("✅ Policy '%s' deleted successfully!\n", policyID)
	}
}

func (cli *PolicyCLI) showPolicyDetails() {
	fmt.Println("\n🔍 Policy Details")
	fmt.Println("=================")

	fmt.Print("Enter policy ID: ")
	policyID := cli.readString()

	if policyID == "" {
		fmt.Println("❌ Policy ID cannot be empty")
		return
	}

	ctx := context.Background()
	policy, err := cli.manager.Get(ctx, policyID)
	if err != nil {
		fmt.Printf("❌ Failed to get policy: %v\n", err)
		return
	}

	fmt.Printf("📋 Policy: %s\n", policy.GetID())
	fmt.Printf("Description: %s\n", policy.GetDescription())
	fmt.Printf("Effect: %s\n", policy.GetEffect())
	fmt.Printf("Subjects: %v\n", policy.GetSubjects())
	fmt.Printf("Resources: %v\n", policy.GetResources())
	fmt.Printf("Actions: %v\n", policy.GetActions())
}

func (cli *PolicyCLI) readString() string {
	input, _ := cli.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
	dbString := os.Getenv("DB_STRING")
	if dbString == "" {
		log.Fatal("DB_STRING environment variable not set")
	}

	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	cli := NewPolicyCLI(db)
	cli.Run()
}
