package models

// Validator interface for models that can be validated
type Validator interface {
	Validate() error
}

// Entity interface for common entity operations
type Entity interface {
	Validator
	TableName() string
}

// PolicyEntity interface for entities that can be associated with policies
type PolicyEntity interface {
	Entity
	GetID() string
}
