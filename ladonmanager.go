package ladonsqlmanager

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ladonsqlmanager/models"
	"github.com/ory/ladon"
	"github.com/ory/ladon/compiler"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ladonsqlmanager/migrations"
)

var (
	// ErrInvalidDriver returned if driver is not postgres or mysql
	ErrInvalidDriver = errors.New("invalid drivername specified, must be mysql or postgres, pg, pgx")
)

// SQLManager implements the ladon/Manager without requiring sqlx or migrations packages
type SQLManager struct {
	db         *gorm.DB
	driverName string
}

// New creates a new, uninitialized SQLManager
func New(db *gorm.DB, driverName string) *SQLManager {
	return &SQLManager{db: db, driverName: driverName}
}

// Init ensures the database is properly initialized with GORM models
func (s *SQLManager) Init() error {
	// Use the migration package to set up the database
	return migrations.Migrate(s.db)
}

// Update updates a policy in the database by deleting original and re-creating
func (s *SQLManager) Update(ctx context.Context, policy ladon.Policy) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.delete(policy.GetID(), tx); err != nil {
			return err
		}
		return s.create(policy, tx)
	})
}

// Create inserts a new policy
func (s *SQLManager) Create(ctx context.Context, policy ladon.Policy) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.create(policy, tx)
	})
}

func (s *SQLManager) create(policy ladon.Policy, tx *gorm.DB) error {
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
		Conditions:  conditions,
		Meta:        meta,
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
	if err := s.processPolicyItems(policy.GetSubjects(), "subject", policy.GetID(), policy.GetStartDelimiter(), policy.GetEndDelimiter(), tx); err != nil {
		return err
	}

	// Process actions
	if err := s.processPolicyItems(policy.GetActions(), "action", policy.GetID(), policy.GetStartDelimiter(), policy.GetEndDelimiter(), tx); err != nil {
		return err
	}

	// Process resources
	if err := s.processPolicyItems(policy.GetResources(), "resource", policy.GetID(), policy.GetStartDelimiter(), policy.GetEndDelimiter(), tx); err != nil {
		return err
	}

	return nil
}

func (s *SQLManager) processPolicyItems(items []string, itemType string, policyID string, startDelim, endDelim byte, tx *gorm.DB) error {
	for _, template := range items {
		h := sha256.New()
		h.Write([]byte(template))
		id := fmt.Sprintf("%x", h.Sum(nil))

		compiled, err := compiler.CompileRegex(template, startDelim, endDelim)
		if err != nil {
			return errors.WithStack(err)
		}

		hasRegex := strings.Index(template, string(startDelim)) >= 0

		// Create or update the item
		var item interface{}
		switch itemType {
		case "subject":
			item = &models.Subject{
				ID:       id,
				Template: template,
				Compiled: compiled.String(),
				HasRegex: hasRegex,
			}
		case "action":
			item = &models.Action{
				ID:       id,
				Template: template,
				Compiled: compiled.String(),
				HasRegex: hasRegex,
			}
		case "resource":
			item = &models.Resource{
				ID:       id,
				Template: template,
				Compiled: compiled.String(),
				HasRegex: hasRegex,
			}
		}

		// Use GORM's FirstOrCreate to avoid duplicates
		if err := tx.Where("id = ?", id).FirstOrCreate(item).Error; err != nil {
			return errors.WithStack(err)
		}

		// Create the relationship
		if err := s.createPolicyRelation(policyID, id, itemType, tx); err != nil {
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

// FindRequestCandidates returns policies that potentially match a ladon.Request
func (s *SQLManager) FindRequestCandidates(ctx context.Context, r *ladon.Request) (ladon.Policies, error) {
	var policies []models.Policy

	// Use GORM to find policies with matching subjects
	query := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Joins("JOIN ladon_policy_subject_rel psr ON psr.policy = ladon_policy.id").
		Joins("JOIN ladon_subject s ON s.id = psr.subject")

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
		if err == gorm.ErrRecordNotFound {
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
	var policy models.Policy

	err := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Where("id = ?", id).
		First(&policy).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
	var policies []models.Policy

	query := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Joins("JOIN ladon_policy_subject_rel psr ON psr.policy = ladon_policy.id").
		Joins("JOIN ladon_subject s ON s.id = psr.subject")

	// Database-specific regex handling
	switch s.driverName {
	case "postgres", "pg", "pgx":
		query = query.Where("(s.has_regex = ? AND ? ~ s.compiled) OR (s.has_regex = ? AND s.template = ?)",
			true, subject, false, subject)
	case "mysql":
		query = query.Where("(s.has_regex = ? AND ? REGEXP BINARY s.compiled) OR (s.has_regex = ? AND s.template = ?)",
			true, subject, false, subject)
	default:
		return nil, ErrInvalidDriver
	}

	err := query.Find(&policies).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}

	return s.convertPoliciesToLadon(policies), nil
}

// FindPoliciesForResource returns policies that could match the resource.
func (s *SQLManager) FindPoliciesForResource(ctx context.Context, resource string) (ladon.Policies, error) {
	var policies []models.Policy

	query := s.db.WithContext(ctx).
		Preload("Subjects").
		Preload("Actions").
		Preload("Resources").
		Joins("JOIN ladon_policy_resource_rel prr ON prr.policy = ladon_policy.id").
		Joins("JOIN ladon_resource r ON r.id = prr.resource")

	// Database-specific regex handling
	switch s.driverName {
	case "postgres", "pg", "pgx":
		query = query.Where("(r.has_regex = ? AND ? ~ r.compiled) OR (r.has_regex = ? AND r.template = ?)",
			true, resource, false, resource)
	case "mysql":
		query = query.Where("(r.has_regex = ? AND ? REGEXP BINARY r.compiled) OR (r.has_regex = ? AND r.template = ?)",
			true, resource, false, resource)
	default:
		return nil, ErrInvalidDriver
	}

	err := query.Find(&policies).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
		json.Unmarshal(policy.Conditions, &ladonPolicy.Conditions)
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
