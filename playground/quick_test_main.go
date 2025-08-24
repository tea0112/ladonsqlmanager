package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ladonsqlmanager"
	"github.com/ory/ladon"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func main() {
	fmt.Println("🧪 Ladon Quick Test - Policy Scenarios")
	fmt.Println("======================================")

	if err := loadConfig("config.env"); err != nil {
		log.Printf("Warning: Could not load config.env: %v", err)
	}

	dbString := os.Getenv("DB_STRING")
	if dbString == "" {
		log.Fatal("DB_STRING environment variable not set. Please set it or create a config.env file.")
	}

	db, err := gorm.Open(postgres.Open(dbString), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	manager := ladonsqlmanager.New(db, "postgres")
	if err := manager.Init(); err != nil {
		log.Fatal("Failed to initialize manager:", err)
	}

	warden := &ladon.Ladon{Manager: manager}
	ctx := context.Background()

	fmt.Println("\n🔐 Test 1: Basic Policy Creation")
	fmt.Println("--------------------------------")

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

	fmt.Println("\n🎉 Quick test completed!")
}
