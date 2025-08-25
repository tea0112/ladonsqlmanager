package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// JSONText is a custom type that can handle JSON stored as text in the database.
// It implements the sql.Scanner, driver.Valuer, json.Marshaler, and json.Unmarshaler interfaces
// to seamlessly handle JSON data between Go structs and database text fields.
//
// Example usage:
//
//	conditions := JSONText(`{"user_role": "admin", "resource_type": "document"}`)
//
//	// Store in database
//	policy := Policy{Conditions: conditions}
//
//	// Retrieve and use
//	if !conditions.IsNull() && conditions.IsValid() {
//		fmt.Println(conditions.String())
//	}
type JSONText json.RawMessage

// Value implements the driver.Valuer interface
func (j JSONText) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

// Scan implements the sql.Scanner interface
func (j *JSONText) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		*j = JSONText(v)
	case []byte:
		*j = JSONText(v)
	default:
		return errors.New("cannot scan non-string value into JSONText")
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (j JSONText) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (j *JSONText) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("cannot unmarshal into nil JSONText")
	}

	// Validate that the data is valid JSON
	if !json.Valid(data) {
		return errors.New("invalid JSON data")
	}

	*j = JSONText(data)
	return nil
}

// IsValid checks if the JSONText contains valid JSON
func (j JSONText) IsValid() bool {
	if j == nil {
		return true // nil is considered valid (represents null)
	}
	return json.Valid([]byte(j))
}

// IsNull checks if the JSONText represents a null value
func (j JSONText) IsNull() bool {
	return j == nil || string(j) == "null"
}

// String returns the string representation of the JSONText
func (j JSONText) String() string {
	if j == nil {
		return "null"
	}
	return string(j)
}

// BaseEntity represents the common structure for Subject, Action, and Resource entities.
// It contains shared fields and behavior for entities that can be associated with policies.
//
// Fields:
//   - ID: Unique identifier for the entity (max 64 chars)
//   - HasRegex: Indicates if the template contains regex patterns
//   - Compiled: Compiled/processed version of the template (max 511 chars)
//   - Template: Original template string (max 511 chars)
//   - CreatedAt: Timestamp when the entity was created
//   - UpdatedAt: Timestamp when the entity was last updated
//   - DeletedAt: Soft delete timestamp (GORM soft delete)
type BaseEntity struct {
	ID        string         `gorm:"column:id;type:varchar(64);primaryKey;not null"`
	HasRegex  bool           `gorm:"column:has_regex;type:bool;not null"`
	Compiled  string         `gorm:"column:compiled;type:varchar(511);uniqueIndex;not null"`
	Template  string         `gorm:"column:template;type:varchar(511);uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// Validate validates the base entity fields
func (b *BaseEntity) Validate() error {
	if b.ID == "" {
		return errors.New("entity ID cannot be empty")
	}
	if len(b.ID) > EntityIDMaxLength {
		return errors.New("entity ID exceeds maximum length")
	}
	if b.Compiled == "" {
		return errors.New("compiled field cannot be empty")
	}
	if len(b.Compiled) > CompiledMaxLength {
		return errors.New("compiled field exceeds maximum length")
	}
	if b.Template == "" {
		return errors.New("template field cannot be empty")
	}
	if len(b.Template) > TemplateMaxLength {
		return errors.New("template field exceeds maximum length")
	}
	return nil
}

// GetID returns the entity ID
func (b *BaseEntity) GetID() string {
	return b.ID
}
