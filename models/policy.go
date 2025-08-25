package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Policy represents the main policy table
type Policy struct {
	ID          string         `gorm:"column:id;type:varchar(255);primaryKey;not null"`
	Description string         `gorm:"column:description;type:text;not null"`
	Effect      string         `gorm:"column:effect;type:text;not null;check:effect IN ('allow', 'deny')"`
	Conditions  JSONText       `gorm:"column:conditions;type:text;not null"`
	Meta        JSONText       `gorm:"column:meta;type:text"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relationships
	Subjects  []Subject  `gorm:"many2many:ladon_policy_subject_rel;foreignKey:ID;joinForeignKey:Policy;References:ID;joinReferences:Subject"`
	Actions   []Action   `gorm:"many2many:ladon_policy_action_rel;foreignKey:ID;joinForeignKey:Policy;References:ID;joinReferences:Action"`
	Resources []Resource `gorm:"many2many:ladon_policy_resource_rel;foreignKey:ID;joinForeignKey:Policy;References:ID;joinReferences:Resource"`
}

// TableName specifies the table name for Policy
func (Policy) TableName() string {
	return TableNamePolicy
}

// Validate validates the policy fields
func (p *Policy) Validate() error {
	if p.ID == "" {
		return errors.New("policy ID cannot be empty")
	}
	if len(p.ID) > PolicyIDMaxLength {
		return errors.New("policy ID exceeds maximum length")
	}
	if p.Description == "" {
		return errors.New("policy description cannot be empty")
	}
	if p.Effect != EffectAllow && p.Effect != EffectDeny {
		return errors.New("effect must be 'allow' or 'deny'")
	}
	if p.Conditions == nil {
		return errors.New("policy conditions cannot be nil")
	}
	return nil
}

// IsAllowEffect returns true if the policy effect is allow
func (p *Policy) IsAllowEffect() bool {
	return p.Effect == EffectAllow
}

// IsDenyEffect returns true if the policy effect is deny
func (p *Policy) IsDenyEffect() bool {
	return p.Effect == EffectDeny
}

// GetID returns the policy ID
func (p *Policy) GetID() string {
	return p.ID
}
