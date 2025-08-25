package ladonsqlmanager

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ladonsqlmanager/migrations"
	"github.com/ladonsqlmanager/models"
	"github.com/ory/ladon"
	"github.com/ory/ladon/compiler"
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
	db         *gorm.DB
	driverName string
	config     Config
}

// New creates a new, uninitialized SQLManager with default configuration
func New(db *gorm.DB, driverName string) *SQLManager {
	return NewWithConfig(db, driverName, DefaultConfig())
}

// NewWithConfig creates a new SQLManager with custom configuration
func NewWithConfig(db *gorm.DB, driverName string, config Config) *SQLManager {
	return &SQLManager{
		db:         db,
		driverName: strings.ToLower(driverName),
		config:     config,
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
	// Batch process items for better performance
	itemEntities := make([]interface{}, 0, len(items))
	relationships := make([]interface{}, 0, len(items))

	for _, template := range items {
		// Sanitize the template
		template = sanitizeTemplate(template)
		if template == "" {
			continue
		}

		h := sha256.New()
		_, _ = h.Write([]byte(template))
		id := fmt.Sprintf("%x", h.Sum(nil))

		compiled, err := compiler.CompileRegex(template, startDelim, endDelim)
		if err != nil {
			return errors.WithStack(err)
		}

		hasRegex := strings.Index(template, string(startDelim)) >= 0

		// Create the base entity
		baseEntity := models.BaseEntity{
			ID:       id,
			Template: template,
			Compiled: compiled.String(),
			HasRegex: hasRegex,
		}

		// Create or update the item
		var item interface{}
		var relation interface{}
		switch itemType {
		case "subject":
			item = &models.Subject{BaseEntity: baseEntity}
			relation = &models.PolicySubjectRel{
				Policy:  policyID,
				Subject: id,
			}
		case "action":
			item = &models.Action{BaseEntity: baseEntity}
			relation = &models.PolicyActionRel{
				Policy: policyID,
				Action: id,
			}
		case "resource":
			item = &models.Resource{BaseEntity: baseEntity}
			relation = &models.PolicyResourceRel{
				Policy:   policyID,
				Resource: id,
			}
		}

		// Use GORM's FirstOrCreate to avoid duplicates
		if err := tx.Where("id = ?", id).FirstOrCreate(item).Error; err != nil {
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
	switch itemType {
	case "subject":
		rel := &models.PolicySubjectRel{
			Policy:  policyID,
			Subject: itemID,
		}
		return tx.Where("policy = ? AND subject = ?", policyID, itemID).FirstOrCreate(rel).Error
	case "action":
		rel := &models.PolicyActionRel{
			Policy: policyID,
			Action: itemID,
		}
		return tx.Where("policy = ? AND action = ?", policyID, itemID).FirstOrCreate(rel).Error
	case "resource":
		rel := &models.PolicyResourceRel{
			Policy:   policyID,
			Resource: itemID,
		}
		return tx.Where("policy = ? AND resource = ?", policyID, itemID).FirstOrCreate(rel).Error
	}
	return nil
}

// createPolicyRelationOptimized creates policy relations with better error handling
func (s *SQLManager) createPolicyRelationOptimized(relation interface{}, tx *gorm.DB) error {
	switch rel := relation.(type) {
	case *models.PolicySubjectRel:
		return tx.Where("policy = ? AND subject = ?", rel.Policy, rel.Subject).FirstOrCreate(rel).Error
	case *models.PolicyActionRel:
		return tx.Where("policy = ? AND action = ?", rel.Policy, rel.Action).FirstOrCreate(rel).Error
	case *models.PolicyResourceRel:
		return tx.Where("policy = ? AND resource = ?", rel.Policy, rel.Resource).FirstOrCreate(rel).Error
	default:
		return errors.New("unsupported relation type")
	}
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

// Helper function to get unique strings (kept for potential future use)
func uniq(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}
