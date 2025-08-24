package migrations

import (
	"fmt"
	"log"

	"github.com/ladonsqlmanager/models"
	"gorm.io/gorm"
)

// Migrate runs database migrations to set up the schema
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.Policy{},
		&models.Subject{},
		&models.Action{},
		&models.Resource{},
		&models.PolicySubjectRel{},
		&models.PolicyActionRel{},
		&models.PolicyResourceRel{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// DropTables drops all tables (useful for testing or resetting)
func DropTables(db *gorm.DB) error {
	log.Println("Dropping all tables...")

	err := db.Migrator().DropTable(
		&models.PolicyResourceRel{},
		&models.PolicyActionRel{},
		&models.PolicySubjectRel{},
		&models.Resource{},
		&models.Action{},
		&models.Subject{},
		&models.Policy{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("All tables dropped successfully")
	return nil
}

// ResetDatabase drops all tables and recreates them
func ResetDatabase(db *gorm.DB) error {
	if err := DropTables(db); err != nil {
		return err
	}
	return Migrate(db)
}
