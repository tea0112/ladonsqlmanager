package models

// Subject represents the subject definitions table
type Subject struct {
	BaseEntity

	// Relationships
	Policies []Policy `gorm:"many2many:ladon_policy_subject_rel;foreignKey:ID;joinForeignKey:Subject;References:ID;joinReferences:Policy"`
}

// TableName specifies the table name for Subject
func (Subject) TableName() string {
	return TableNameSubject
}

// Validate validates the subject
func (s *Subject) Validate() error {
	return s.BaseEntity.Validate()
}

// Action represents the action definitions table
type Action struct {
	BaseEntity

	// Relationships
	Policies []Policy `gorm:"many2many:ladon_policy_action_rel;foreignKey:ID;joinForeignKey:Action;References:ID;joinReferences:Policy"`
}

// TableName specifies the table name for Action
func (Action) TableName() string {
	return TableNameAction
}

// Validate validates the action
func (a *Action) Validate() error {
	return a.BaseEntity.Validate()
}

// Resource represents the resource definitions table
type Resource struct {
	BaseEntity

	// Relationships
	Policies []Policy `gorm:"many2many:ladon_policy_resource_rel;foreignKey:ID;joinForeignKey:Resource;References:ID;joinReferences:Policy"`
}

// TableName specifies the table name for Resource
func (Resource) TableName() string {
	return TableNameResource
}

// Validate validates the resource
func (r *Resource) Validate() error {
	return r.BaseEntity.Validate()
}
