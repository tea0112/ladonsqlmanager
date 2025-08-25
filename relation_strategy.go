package ladonsqlmanager

import (
	"github.com/ladonsqlmanager/models"
	"gorm.io/gorm"
)

// RelationStrategy defines the interface for handling different types of policy relations
type RelationStrategy interface {
	// CreateRelation creates a new relation instance
	CreateRelation(policyID, entityID string) interface{}
	// PersistRelation persists the relation to the database using GORM
	PersistRelation(relation interface{}, tx *gorm.DB) error
	// GetRelationType returns the type identifier for this relation
	GetRelationType() string
}

// SubjectRelationStrategy handles PolicySubjectRel operations
type SubjectRelationStrategy struct{}

// CreateRelation creates a new PolicySubjectRel
func (s *SubjectRelationStrategy) CreateRelation(policyID, entityID string) interface{} {
	return &models.PolicySubjectRel{
		Policy:  policyID,
		Subject: entityID,
	}
}

// PersistRelation persists a PolicySubjectRel to the database
func (s *SubjectRelationStrategy) PersistRelation(relation interface{}, tx *gorm.DB) error {
	rel, ok := relation.(*models.PolicySubjectRel)
	if !ok {
		return ErrInvalidRelationType
	}
	return tx.Where("policy = ? AND subject = ?", rel.Policy, rel.Subject).FirstOrCreate(rel).Error
}

// GetRelationType returns the relation type identifier
func (s *SubjectRelationStrategy) GetRelationType() string {
	return itemTypeSubject
}

// ActionRelationStrategy handles PolicyActionRel operations
type ActionRelationStrategy struct{}

// CreateRelation creates a new PolicyActionRel
func (a *ActionRelationStrategy) CreateRelation(policyID, entityID string) interface{} {
	return &models.PolicyActionRel{
		Policy: policyID,
		Action: entityID,
	}
}

// PersistRelation persists a PolicyActionRel to the database
func (a *ActionRelationStrategy) PersistRelation(relation interface{}, tx *gorm.DB) error {
	rel, ok := relation.(*models.PolicyActionRel)
	if !ok {
		return ErrInvalidRelationType
	}
	return tx.Where("policy = ? AND action = ?", rel.Policy, rel.Action).FirstOrCreate(rel).Error
}

// GetRelationType returns the relation type identifier
func (a *ActionRelationStrategy) GetRelationType() string {
	return itemTypeAction
}

// ResourceRelationStrategy handles PolicyResourceRel operations
type ResourceRelationStrategy struct{}

// CreateRelation creates a new PolicyResourceRel
func (r *ResourceRelationStrategy) CreateRelation(policyID, entityID string) interface{} {
	return &models.PolicyResourceRel{
		Policy:   policyID,
		Resource: entityID,
	}
}

// PersistRelation persists a PolicyResourceRel to the database
func (r *ResourceRelationStrategy) PersistRelation(relation interface{}, tx *gorm.DB) error {
	rel, ok := relation.(*models.PolicyResourceRel)
	if !ok {
		return ErrInvalidRelationType
	}
	return tx.Where("policy = ? AND resource = ?", rel.Policy, rel.Resource).FirstOrCreate(rel).Error
}

// GetRelationType returns the relation type identifier
func (r *ResourceRelationStrategy) GetRelationType() string {
	return itemTypeResource
}

// RelationStrategyRegistry manages the available relation strategies
type RelationStrategyRegistry struct {
	strategies map[string]RelationStrategy
}

// NewRelationStrategyRegistry creates a new registry with default strategies
func NewRelationStrategyRegistry() *RelationStrategyRegistry {
	registry := &RelationStrategyRegistry{
		strategies: make(map[string]RelationStrategy),
	}

	// Register default strategies
	registry.RegisterStrategy(itemTypeSubject, &SubjectRelationStrategy{})
	registry.RegisterStrategy(itemTypeAction, &ActionRelationStrategy{})
	registry.RegisterStrategy(itemTypeResource, &ResourceRelationStrategy{})

	return registry
}

// RegisterStrategy registers a new strategy for the given relation type
func (r *RelationStrategyRegistry) RegisterStrategy(relationType string, strategy RelationStrategy) {
	r.strategies[relationType] = strategy
}

// GetStrategy returns the strategy for the given relation type
func (r *RelationStrategyRegistry) GetStrategy(relationType string) (RelationStrategy, bool) {
	strategy, exists := r.strategies[relationType]
	return strategy, exists
}

// GetSupportedTypes returns all supported relation types
func (r *RelationStrategyRegistry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.strategies))
	for relationType := range r.strategies {
		types = append(types, relationType)
	}
	return types
}

// RelationContext provides context for relation operations
type RelationContext struct {
	strategy RelationStrategy
}

// NewRelationContext creates a new context with the given strategy
func NewRelationContext(strategy RelationStrategy) *RelationContext {
	return &RelationContext{strategy: strategy}
}

// SetStrategy changes the strategy at runtime
func (c *RelationContext) SetStrategy(strategy RelationStrategy) {
	c.strategy = strategy
}

// CreateRelation creates a relation using the current strategy
func (c *RelationContext) CreateRelation(policyID, entityID string) interface{} {
	return c.strategy.CreateRelation(policyID, entityID)
}

// PersistRelation persists a relation using the current strategy
func (c *RelationContext) PersistRelation(relation interface{}, tx *gorm.DB) error {
	return c.strategy.PersistRelation(relation, tx)
}

// RelationTypeDetector provides methods to detect relation types
type RelationTypeDetector struct {
	strategyRegistry *RelationStrategyRegistry
}

// NewRelationTypeDetector creates a new detector with the given registry
func NewRelationTypeDetector(registry *RelationStrategyRegistry) *RelationTypeDetector {
	return &RelationTypeDetector{strategyRegistry: registry}
}

// DetectAndGetStrategy detects the relation type and returns the appropriate strategy
func (d *RelationTypeDetector) DetectAndGetStrategy(relation interface{}) (RelationStrategy, error) {
	switch relation.(type) {
	case *models.PolicySubjectRel:
		if strategy, exists := d.strategyRegistry.GetStrategy(itemTypeSubject); exists {
			return strategy, nil
		}
	case *models.PolicyActionRel:
		if strategy, exists := d.strategyRegistry.GetStrategy(itemTypeAction); exists {
			return strategy, nil
		}
	case *models.PolicyResourceRel:
		if strategy, exists := d.strategyRegistry.GetStrategy(itemTypeResource); exists {
			return strategy, nil
		}
	}
	return nil, ErrInvalidRelationType
}
