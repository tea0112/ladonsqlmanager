package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Policy represents the main policy table
type Policy struct {
	ID          string          `gorm:"column:id;type:varchar(255);primaryKey;not null"`
	Description string          `gorm:"column:description;type:text;not null"`
	Effect      string          `gorm:"column:effect;type:text;not null;check:effect IN ('allow', 'deny')"`
	Conditions  json.RawMessage `gorm:"column:conditions;type:text;not null"`
	Meta        json.RawMessage `gorm:"column:meta;type:text"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt  `gorm:"column:deleted_at;index"`

	// Relationships
	Subjects  []Subject  `gorm:"many2many:ladon_policy_subject_rel;foreignKey:ID;joinForeignKey:Policy;References:ID;joinReferences:Subject"`
	Actions   []Action   `gorm:"many2many:ladon_policy_action_rel;foreignKey:ID;joinForeignKey:Policy;References:ID;joinReferences:Action"`
	Resources []Resource `gorm:"many2many:ladon_policy_resource_rel;foreignKey:ID;joinForeignKey:Policy;References:ID;joinReferences:Resource"`
}

// TableName specifies the table name for Policy
func (Policy) TableName() string {
	return "ladon_policy"
}

// Subject represents the subject definitions table
type Subject struct {
	ID        string         `gorm:"column:id;type:varchar(64);primaryKey;not null"`
	HasRegex  bool           `gorm:"column:has_regex;type:bool;not null"`
	Compiled  string         `gorm:"column:compiled;type:varchar(511);uniqueIndex;not null"`
	Template  string         `gorm:"column:template;type:varchar(511);uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	Policies []Policy `gorm:"many2many:ladon_policy_subject_rel;foreignKey:ID;joinForeignKey:Subject;References:ID;joinReferences:Policy"`
}

// TableName specifies the table name for Subject
func (Subject) TableName() string {
	return "ladon_subject"
}

// Action represents the action definitions table
type Action struct {
	ID        string         `gorm:"column:id;type:varchar(64);primaryKey;not null"`
	HasRegex  bool           `gorm:"column:has_regex;type:bool;not null"`
	Compiled  string         `gorm:"column:compiled;type:varchar(511);uniqueIndex;not null"`
	Template  string         `gorm:"column:template;type:varchar(511);uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	Policies []Policy `gorm:"many2many:ladon_policy_action_rel;foreignKey:ID;joinForeignKey:Action;References:ID;joinReferences:Policy"`
}

// TableName specifies the table name for Action
func (Action) TableName() string {
	return "ladon_action"
}

// Resource represents the resource definitions table
type Resource struct {
	ID        string         `gorm:"column:id;type:varchar(64);primaryKey;not null"`
	HasRegex  bool           `gorm:"column:has_regex;type:bool;not null"`
	Compiled  string         `gorm:"column:compiled;type:varchar(511);uniqueIndex;not null"`
	Template  string         `gorm:"column:template;type:varchar(511);uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	Policies []Policy `gorm:"many2many:ladon_policy_resource_rel;foreignKey:ID;joinForeignKey:Resource;References:ID;joinReferences:Policy"`
}

// TableName specifies the table name for Resource
func (Resource) TableName() string {
	return "ladon_resource"
}

// PolicySubjectRel represents the policy-subject relationship table
type PolicySubjectRel struct {
	Policy    string    `gorm:"column:policy;type:varchar(255);primaryKey;not null"`
	Subject   string    `gorm:"column:subject;type:varchar(64);primaryKey;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`

	// Foreign key relationships
	PolicyRef  Policy  `gorm:"foreignKey:Policy;references:ID;constraint:OnDelete:CASCADE"`
	SubjectRef Subject `gorm:"foreignKey:Subject;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for PolicySubjectRel
func (PolicySubjectRel) TableName() string {
	return "ladon_policy_subject_rel"
}

// PolicyActionRel represents the policy-action relationship table
type PolicyActionRel struct {
	Policy    string    `gorm:"column:policy;type:varchar(255);primaryKey;not null"`
	Action    string    `gorm:"column:action;type:varchar(64);primaryKey;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`

	// Foreign key relationships
	PolicyRef Policy `gorm:"foreignKey:Policy;references:ID;constraint:OnDelete:CASCADE"`
	ActionRef Action `gorm:"foreignKey:Action;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for PolicyActionRel
func (PolicyActionRel) TableName() string {
	return "ladon_policy_action_rel"
}

// PolicyResourceRel represents the policy-resource relationship table
type PolicyResourceRel struct {
	Policy    string    `gorm:"column:policy;type:varchar(255);primaryKey;not null"`
	Resource  string    `gorm:"column:resource;type:varchar(64);primaryKey;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`

	// Foreign key relationships
	PolicyRef   Policy   `gorm:"foreignKey:Policy;references:ID;constraint:OnDelete:CASCADE"`
	ResourceRef Resource `gorm:"foreignKey:Resource;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for PolicyResourceRel
func (PolicyResourceRel) TableName() string {
	return "ladon_policy_resource_rel"
}
