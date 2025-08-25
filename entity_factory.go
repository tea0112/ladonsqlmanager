package ladonsqlmanager

import (
	"github.com/ladonsqlmanager/models"
)

// EntityFactory defines the interface for creating entities and their relationships
type EntityFactory interface {
	CreateEntity(baseEntity models.BaseEntity) interface{}
	CreateRelation(policyID, entityID string) interface{}
	GetEntityType() string
	GetRelationStrategy() RelationStrategy
}

// SubjectFactory creates Subject entities and PolicySubjectRel relationships
type SubjectFactory struct {
	relationStrategy RelationStrategy
}

// CreateEntity creates a new Subject entity
func (f *SubjectFactory) CreateEntity(baseEntity models.BaseEntity) interface{} {
	return &models.Subject{BaseEntity: baseEntity}
}

// CreateRelation creates a new PolicySubjectRel relationship
func (f *SubjectFactory) CreateRelation(policyID, entityID string) interface{} {
	return &models.PolicySubjectRel{
		Policy:  policyID,
		Subject: entityID,
	}
}

// GetEntityType returns the entity type identifier
func (f *SubjectFactory) GetEntityType() string {
	return itemTypeSubject
}

// GetRelationStrategy returns the relation strategy for this factory
func (f *SubjectFactory) GetRelationStrategy() RelationStrategy {
	if f.relationStrategy == nil {
		f.relationStrategy = &SubjectRelationStrategy{}
	}
	return f.relationStrategy
}

// ActionFactory creates Action entities and PolicyActionRel relationships
type ActionFactory struct {
	relationStrategy RelationStrategy
}

// CreateEntity creates a new Action entity
func (f *ActionFactory) CreateEntity(baseEntity models.BaseEntity) interface{} {
	return &models.Action{BaseEntity: baseEntity}
}

// CreateRelation creates a new PolicyActionRel relationship
func (f *ActionFactory) CreateRelation(policyID, entityID string) interface{} {
	return &models.PolicyActionRel{
		Policy: policyID,
		Action: entityID,
	}
}

// GetEntityType returns the entity type identifier
func (f *ActionFactory) GetEntityType() string {
	return itemTypeAction
}

// GetRelationStrategy returns the relation strategy for this factory
func (f *ActionFactory) GetRelationStrategy() RelationStrategy {
	if f.relationStrategy == nil {
		f.relationStrategy = &ActionRelationStrategy{}
	}
	return f.relationStrategy
}

// ResourceFactory creates Resource entities and PolicyResourceRel relationships
type ResourceFactory struct {
	relationStrategy RelationStrategy
}

// CreateEntity creates a new Resource entity
func (f *ResourceFactory) CreateEntity(baseEntity models.BaseEntity) interface{} {
	return &models.Resource{BaseEntity: baseEntity}
}

// CreateRelation creates a new PolicyResourceRel relationship
func (f *ResourceFactory) CreateRelation(policyID, entityID string) interface{} {
	return &models.PolicyResourceRel{
		Policy:   policyID,
		Resource: entityID,
	}
}

// GetEntityType returns the entity type identifier
func (f *ResourceFactory) GetEntityType() string {
	return itemTypeResource
}

// GetRelationStrategy returns the relation strategy for this factory
func (f *ResourceFactory) GetRelationStrategy() RelationStrategy {
	if f.relationStrategy == nil {
		f.relationStrategy = &ResourceRelationStrategy{}
	}
	return f.relationStrategy
}

// EntityFactoryRegistry manages the available entity factories
type EntityFactoryRegistry struct {
	factories map[string]EntityFactory
}

// NewEntityFactoryRegistry creates a new registry with default factories
func NewEntityFactoryRegistry() *EntityFactoryRegistry {
	registry := &EntityFactoryRegistry{
		factories: make(map[string]EntityFactory),
	}

	// Register default factories
	registry.RegisterFactory(itemTypeSubject, &SubjectFactory{})
	registry.RegisterFactory(itemTypeAction, &ActionFactory{})
	registry.RegisterFactory(itemTypeResource, &ResourceFactory{})

	return registry
}

// RegisterFactory registers a new factory for the given entity type
func (r *EntityFactoryRegistry) RegisterFactory(entityType string, factory EntityFactory) {
	r.factories[entityType] = factory
}

// GetFactory returns the factory for the given entity type
func (r *EntityFactoryRegistry) GetFactory(entityType string) (EntityFactory, bool) {
	factory, exists := r.factories[entityType]
	return factory, exists
}

// GetSupportedTypes returns all supported entity types
func (r *EntityFactoryRegistry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.factories))
	for entityType := range r.factories {
		types = append(types, entityType)
	}
	return types
}
