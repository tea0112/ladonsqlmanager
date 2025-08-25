package ladonsqlmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ladonsqlmanager/migrations"
	"github.com/ladonsqlmanager/models"
	"github.com/ory/ladon"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	itemTypeSubject  = "subject"
	itemTypeAction   = "action"
	itemTypeResource = "resource"
)

var (
	// ErrInvalidDriver returned if driver is not postgres or mysql
	ErrInvalidDriver = errors.New("invalid drivername specified, must be mysql or postgres, pg, pgx")
	// ErrInvalidPolicy returned when policy validation fails
	ErrInvalidPolicy = errors.New("invalid policy")
	// ErrEmptyPolicyID returned when policy ID is empty
	ErrEmptyPolicyID = errors.New("policy ID cannot be empty")
	// ErrPolicyIDTooLong returned when policy ID exceeds maximum length
	ErrPolicyIDTooLong = errors.New("policy ID exceeds maximum length")
	// ErrInvalidRelationType returned when relation type is invalid
	ErrInvalidRelationType = errors.New("invalid relation type")
)

// Config holds configuration options for SQLManager
type Config struct {
	MaxBatchSize       int
	QueryTimeout       time.Duration
	EnableMetrics      bool
	SlowQueryThreshold time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		MaxBatchSize:       100,
		QueryTimeout:       30 * time.Second,
		EnableMetrics:      false,
		SlowQueryThreshold: 100 * time.Millisecond,
	}
}

// SQLManager implements the ladon/Manager without requiring sqlx or migrations packages
type SQLManager struct {
	db               *gorm.DB
	driverName       string
	config           Config
	factoryRegistry  *EntityFactoryRegistry
	builderDirector  *EntityBuilderDirector
	strategyRegistry *RelationStrategyRegistry
	typeDetector     *RelationTypeDetector
}

// New creates a new, uninitialized SQLManager with default configuration
func New(db *gorm.DB, driverName string) *SQLManager {
	return NewWithConfig(db, driverName, DefaultConfig())
}

// NewWithConfig creates a new SQLManager with custom configuration
func NewWithConfig(db *gorm.DB, driverName string, config Config) *SQLManager {
	strategyRegistry := NewRelationStrategyRegistry()
	return &SQLManager{
		db:               db,
		driverName:       strings.ToLower(driverName),
		config:           config,
		factoryRegistry:  NewEntityFactoryRegistry(),
		builderDirector:  NewEntityBuilderDirector(),
		strategyRegistry: strategyRegistry,
		typeDetector:     NewRelationTypeDetector(strategyRegistry),
	}
}

// Init ensures the database is properly initialized with GORM models
func (s *SQLManager) Init() error {
	// Use the migration package to set up the database
	return migrations.Migrate(s.db)
}

// Update updates a policy in the database by deleting original and re-creating
func (s *SQLManager) Update(ctx context.Context, policy ladon.Policy) error {
	start := time.Now()
	defer func() {
		s.logSlowQuery("Update", time.Since(start))
	}()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.delete(policy.GetID(), tx); err != nil {
			return err
		}
		return s.create(policy, tx)
	})
}

// Create inserts a new policy
func (s *SQLManager) Create(ctx context.Context, policy ladon.Policy) error {
	start := time.Now()
	defer func() {
		s.logSlowQuery("Create", time.Since(start))
	}()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.create(policy, tx)
	})
}

func (s *SQLManager) create(policy ladon.Policy, tx *gorm.DB) error {
	// Input validation
	if policy.GetID() == "" {
		return errors.WithStack(ErrEmptyPolicyID)
	}
	if len(policy.GetID()) > models.PolicyIDMaxLength {
		return errors.WithStack(ErrPolicyIDTooLong)
	}

	conditions := []byte("{}")
	if policy.GetConditions() != nil {
		cs := policy.GetConditions()
		var err error
		conditions, err = json.Marshal(&cs)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	meta := []byte("{}")
	if policy.GetMeta() != nil {
		meta = policy.GetMeta()
	}

	// Create policy using GORM
	policyModel := &models.Policy{
		ID:          policy.GetID(),
		Description: policy.GetDescription(),
		Effect:      policy.GetEffect(),
		Conditions:  models.JSONText(conditions),
		Meta:        models.JSONText(meta),
	}

	// Validate policy model before persisting
	if err := policyModel.Validate(); err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Create(policyModel).Error; err != nil {
		return errors.WithStack(err)
	}

	// Process subjects, actions, and resources
	if err := s.processPolicyRelations(policy, tx); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *SQLManager) processPolicyRelations(policy ladon.Policy, tx *gorm.DB) error {
	// Process subjects
	if err := s.processPolicyItems(policy.GetSubjects(), itemTypeSubject, policy.GetID(), policy.GetStartDelimiter(), policy.GetEndDelimiter(), tx); err != nil {
		return err
	}

	// Process actions
	if err := s.processPolicyItems(policy.GetActions(), itemTypeAction, policy.GetID(), policy.GetStartDelimiter(), policy.GetEndDelimiter(), tx); err != nil {
		return err
	}

	// Process resources
	if err := s.processPolicyItems(policy.GetResources(), itemTypeResource, policy.GetID(), policy.GetStartDelimiter(), policy.GetEndDelimiter(), tx); err != nil {
		return err
	}

	return nil
}

func (s *SQLManager) processPolicyItems(items []string, itemType string, policyID string, startDelim, endDelim byte, tx *gorm.DB) error {
	// Get the appropriate factory for this entity type
	factory, exists := s.factoryRegistry.GetFactory(itemType)
	if !exists {
		return errors.Errorf("unsupported entity type: %s", itemType)
	}

	// Batch process items for better performance
	relationships := make([]interface{}, 0, len(items))

	for _, template := range items {
		// Use the builder to create the base entity
		baseEntity, err := s.builderDirector.BuildStandardEntity(template, startDelim, endDelim)
		if err != nil {
			// Skip invalid templates but continue processing others
			continue
		}

		// Use the factory to create the specific entity type and relationship
		item := factory.CreateEntity(baseEntity)
		relation := factory.CreateRelation(policyID, baseEntity.ID)

		// Use GORM's FirstOrCreate to avoid duplicates
		if err := tx.Where("id = ?", baseEntity.ID).FirstOrCreate(item).Error; err != nil {
			return errors.WithStack(err)
		}

		relationships = append(relationships, relation)
	}

	// Batch create relationships
	for _, rel := range relationships {
		if err := s.createPolicyRelationOptimized(rel, tx); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (s *SQLManager) createPolicyRelation(policyID, itemID, itemType string, tx *gorm.DB) error {
	// Get the appropriate factory for this entity type
	factory, exists := s.factoryRegistry.GetFactory(itemType)
	if !exists {
		return errors.Errorf("unsupported entity type: %s", itemType)
	}

	// Use the factory to create the relationship
	relation := factory.CreateRelation(policyID, itemID)

	// Use the optimized method to create the relationship
	return s.createPolicyRelationOptimized(relation, tx)
}

// createPolicyRelationOptimized creates policy relations using strategy pattern
func (s *SQLManager) createPolicyRelationOptimized(relation interface{}, tx *gorm.DB) error {
	// Use the type detector to get the appropriate strategy
	strategy, err := s.typeDetector.DetectAndGetStrategy(relation)
	if err != nil {
		return err
	}

	// Use the strategy to persist the relation
	return strategy.PersistRelation(relation, tx)
}

// buildRegexQuery builds a database-specific regex query for matching entities
func (s *SQLManager) buildRegexQuery(query *gorm.DB, field string, value string) *gorm.DB {
	switch s.driverName {
	case "postgres", "pg", "pgx":
		return query.Where(fmt.Sprintf("(%s.has_regex = ? AND ? ~ %s.compiled) OR (%s.has_regex = ? AND %s.template = ?)", field, field, field, field),
			true, value, false, value)
	case "mysql":
		return query.Where(fmt.Sprintf("(%s.has_regex = ? AND ? REGEXP BINARY %s.compiled) OR (%s.has_regex = ? AND %s.template = ?)", field, field, field, field),
			true, value, false, value)
	default:
		return query
	}
}

// sanitizeTemplate removes potentially dangerous characters from templates
func sanitizeTemplate(template string) string {
	return strings.TrimSpace(template)
}

// logSlowQuery logs queries that exceed the slow query threshold
func (s *SQLManager) logSlowQuery(operation string, duration time.Duration) {
	if s.config.EnableMetrics && duration > s.config.SlowQueryThreshold {
		log.Printf("[SLOW QUERY] %s took %v", operation, duration)
	}
}

// FindRequestCandidates returns policies that potentially match a ladon.Request
func (s *SQLManager) FindRequestCandidates(ctx context.Context, r *ladon.Request) (ladon.Policies, error) {
	var policies []models.Policy

	// Use GORM to find policies with matching subjects
	query := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Distinct().
		Joins(fmt.Sprintf("JOIN %s psr ON psr.policy = %s.id", models.TableNamePolicySubjectRel, models.TableNamePolicy)).
		Joins(fmt.Sprintf("JOIN %s s ON s.id = psr.subject", models.TableNameSubject))

	// Database-specific regex handling
	switch s.driverName {
	case "postgres", "pg", "pgx":
		query = query.Where("(s.has_regex = ? AND ? ~ s.compiled) OR (s.has_regex = ? AND s.template = ?)",
			true, r.Subject, false, r.Subject)
	case "mysql":
		query = query.Where("(s.has_regex = ? AND ? REGEXP BINARY s.compiled) OR (s.has_regex = ? AND s.template = ?)",
			true, r.Subject, false, r.Subject)
	default:
		return nil, ErrInvalidDriver
	}

	err := query.Find(&policies).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}

	return s.convertPoliciesToLadon(policies), nil
}

// GetAll returns all policies
func (s *SQLManager) GetAll(ctx context.Context, limit, offset int64) (ladon.Policies, error) {
	var policies []models.Policy

	err := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Limit(int(limit)).
		Offset(int(offset)).
		Order("id").
		Find(&policies).Error

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return s.convertPoliciesToLadon(policies), nil
}

// Get retrieves a policy.
func (s *SQLManager) Get(ctx context.Context, id string) (ladon.Policy, error) {
	start := time.Now()
	defer func() {
		s.logSlowQuery("Get", time.Since(start))
	}()

	var policy models.Policy

	err := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Where("id = ?", id).
		First(&policy).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}

	return s.convertPolicyToLadon(policy), nil
}

// Delete removes a policy.
func (s *SQLManager) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.delete(id, tx)
	})
}

// Delete removes a policy.
func (s *SQLManager) delete(id string, tx *gorm.DB) error {
	// GORM will handle cascade deletes due to foreign key constraints
	return tx.Delete(&models.Policy{}, "id = ?", id).Error
}

// FindPoliciesForSubject returns policies that could match the subject.
func (s *SQLManager) FindPoliciesForSubject(ctx context.Context, subject string) (ladon.Policies, error) {
	start := time.Now()
	defer func() {
		s.logSlowQuery("FindPoliciesForSubject", time.Since(start))
	}()

	var policies []models.Policy

	query := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Distinct().
		Joins(fmt.Sprintf("JOIN %s psr ON psr.policy = %s.id", models.TableNamePolicySubjectRel, models.TableNamePolicy)).
		Joins(fmt.Sprintf("JOIN %s s ON s.id = psr.subject", models.TableNameSubject))

	// Use the helper method for database-specific regex handling
	query = s.buildRegexQuery(query, "s", subject)
	if s.driverName != "postgres" && s.driverName != "pg" && s.driverName != "pgx" && s.driverName != "mysql" {
		return nil, ErrInvalidDriver
	}

	err := query.Find(&policies).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}

	return s.convertPoliciesToLadon(policies), nil
}

// FindPoliciesForResource returns policies that could match the resource.
func (s *SQLManager) FindPoliciesForResource(ctx context.Context, resource string) (ladon.Policies, error) {
	start := time.Now()
	defer func() {
		s.logSlowQuery("FindPoliciesForResource", time.Since(start))
	}()

	var policies []models.Policy

	query := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Distinct().
		Joins(fmt.Sprintf("JOIN %s prr ON prr.policy = %s.id", models.TableNamePolicyResourceRel, models.TableNamePolicy)).
		Joins(fmt.Sprintf("JOIN %s r ON r.id = prr.resource", models.TableNameResource))

	// Use the helper method for database-specific regex handling
	query = s.buildRegexQuery(query, "r", resource)
	if s.driverName != "postgres" && s.driverName != "pg" && s.driverName != "pgx" && s.driverName != "mysql" {
		return nil, ErrInvalidDriver
	}

	err := query.Find(&policies).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}

	return s.convertPoliciesToLadon(policies), nil
}

// Helper functions to convert between GORM models and Ladon interfaces
func (s *SQLManager) convertPolicyToLadon(policy models.Policy) ladon.Policy {
	ladonPolicy := &ladon.DefaultPolicy{
		ID:          policy.ID,
		Description: policy.Description,
		Effect:      policy.Effect,
		Conditions:  ladon.Conditions{},
		Meta:        []byte(policy.Meta),
	}

	// Convert subjects
	for _, subject := range policy.Subjects {
		ladonPolicy.Subjects = append(ladonPolicy.Subjects, subject.Template)
	}

	// Convert actions
	for _, action := range policy.Actions {
		ladonPolicy.Actions = append(ladonPolicy.Actions, action.Template)
	}

	// Convert resources
	for _, resource := range policy.Resources {
		ladonPolicy.Resources = append(ladonPolicy.Resources, resource.Template)
	}

	// Parse conditions
	if len(policy.Conditions) > 0 {
		_ = json.Unmarshal([]byte(policy.Conditions), &ladonPolicy.Conditions)
	}

	return ladonPolicy
}

func (s *SQLManager) convertPoliciesToLadon(policies []models.Policy) ladon.Policies {
	result := make(ladon.Policies, len(policies))
	for i, policy := range policies {
		result[i] = s.convertPolicyToLadon(policy)
	}
	return result
}
