package models

import (
	"time"
)

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
	return TableNamePolicySubjectRel
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
	return TableNamePolicyActionRel
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
	return TableNamePolicyResourceRel
}
