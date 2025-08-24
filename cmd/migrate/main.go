package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ladonsqlmanager/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// loadConfig loads environment variables from config.env file
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
	var (
		action   = flag.String("action", "migrate", "Action to perform: migrate, drop, reset")
		dbString = flag.String("db", "", "Database connection string (overrides config.env)")
		help     = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("Ladon SQL Manager - Migration Tool")
		fmt.Println("==================================")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate/main.go [flags]")
		fmt.Println("")
		fmt.Println("Flags:")
		fmt.Println("  -action string")
		fmt.Println("        Action to perform: migrate, drop, reset (default: migrate)")
		fmt.Println("  -db string")
		fmt.Println("        Database connection string (overrides config.env)")
		fmt.Println("  -help")
		fmt.Println("        Show this help message")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/migrate/main.go -action=migrate")
		fmt.Println("  go run cmd/migrate/main.go -action=drop -db=\"postgres://user:pass@localhost/db\"")
		fmt.Println("")
		fmt.Println("The tool will automatically read from config.env if no -db flag is provided.")
		os.Exit(0)
	}

	// Try to load config.env if no explicit database string is provided
	if *dbString == "" {
		if err := loadConfig("config.env"); err != nil {
			log.Printf("Warning: Could not load config.env: %v", err)
		}

		// Get database string from environment
		*dbString = os.Getenv("DB_STRING")
		if *dbString == "" {
			log.Fatal("Database connection string is required. Use -db flag or set DB_STRING in config.env")
		}
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(*dbString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Perform the requested action
	switch *action {
	case "migrate":
		log.Println("Running database migrations...")
		if err := migrations.Migrate(db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✅ Database migrations completed successfully!")

	case "drop":
		log.Println("Dropping all tables...")
		if err := migrations.DropTables(db); err != nil {
			log.Fatalf("Drop tables failed: %v", err)
		}
		log.Println("✅ All tables dropped successfully!")

	case "reset":
		log.Println("Resetting database...")
		if err := migrations.ResetDatabase(db); err != nil {
			log.Fatalf("Reset database failed: %v", err)
		}
		log.Println("✅ Database reset completed successfully!")

	default:
		log.Fatalf("Unknown action: %s. Valid actions are: migrate, drop, reset", *action)
	}
}
