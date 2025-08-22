package ladonsqlmanager

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ory/ladon"
	"github.com/ory/ladon/compiler"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	// ErrInvalidDriver returned if driver is not postgres or mysql
	ErrInvalidDriver = errors.New("invalid drivername specified, must be mysql or postgres, pg, pgx")
)

// SQLManager implements the ladon/Manager without requiring sqlx or migrations packages
type SQLManager struct {
	db         *gorm.DB
	driverName string
	stmts      *Statements
}

// New creates a new, uninitialized SQLManager
func New(db *gorm.DB, driverName string) *SQLManager {
	return &SQLManager{db: db, driverName: driverName}
}

// SetStatements allows callers to just provide their own statements if they
// want to support something other than postgres/mysql
// Note you must call this before Init() if you wish to override the driver specific
// statements.
func (s *SQLManager) SetStatements(statements *Statements) {
	s.stmts = statements
}

// Init ensures statements are properly mapped
func (s *SQLManager) Init() error {
	if s.stmts == nil {
		s.stmts = GetStatements(s.driverName)
		if s.stmts == nil {
			return ErrInvalidDriver
		}
	}
	return nil
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

	// Insert policy using GORM
	policyData := map[string]interface{}{
		"id":          policy.GetID(),
		"description": policy.GetDescription(),
		"effect":      policy.GetEffect(),
		"conditions":  conditions,
		"meta":        meta,
	}

	if err := tx.Table("ladon_policy").Create(policyData).Error; err != nil {
		return errors.WithStack(err)
	}

	type relation struct {
		p []string
		t string
	}
	var relations = []relation{
		{p: policy.GetActions(), t: "action"},
		{p: policy.GetResources(), t: "resource"},
		{p: policy.GetSubjects(), t: "subject"},
	}

	for _, rel := range relations {
		var query string
		var queryRel string

		switch rel.t {
		case "action":
			query = s.stmts.QueryInsertPolicyActions
			queryRel = s.stmts.QueryInsertPolicyActionsRel
		case "resource":
			query = s.stmts.QueryInsertPolicyResources
			queryRel = s.stmts.QueryInsertPolicyResourcesRel
		case "subject":
			query = s.stmts.QueryInsertPolicySubjects
			queryRel = s.stmts.QueryInsertPolicySubjectsRel
		}

		for _, template := range rel.p {
			h := sha256.New()
			h.Write([]byte(template))
			id := fmt.Sprintf("%x", h.Sum(nil))

			compiled, err := compiler.CompileRegex(template, policy.GetStartDelimiter(), policy.GetEndDelimiter())
			if err != nil {
				return errors.WithStack(err)
			}

			// Use GORM for insertions
			switch rel.t {
			case "action":
				if err := tx.Exec(query, id, template, compiled.String(), strings.Index(template, string(policy.GetStartDelimiter())) >= -1).Error; err != nil {
					return errors.WithStack(err)
				}
				if err := tx.Exec(queryRel, policy.GetID(), id).Error; err != nil {
					return errors.WithStack(err)
				}
			case "resource":
				if err := tx.Exec(query, id, template, compiled.String(), strings.Index(template, string(policy.GetStartDelimiter())) >= -1).Error; err != nil {
					return errors.WithStack(err)
				}
				if err := tx.Exec(queryRel, policy.GetID(), id).Error; err != nil {
					return errors.WithStack(err)
				}
			case "subject":
				if err := tx.Exec(query, id, template, compiled.String(), strings.Index(template, string(policy.GetStartDelimiter())) >= -1).Error; err != nil {
					return errors.WithStack(err)
				}
				if err := tx.Exec(queryRel, policy.GetID(), id).Error; err != nil {
					return errors.WithStack(err)
				}
			}
		}
	}

	return nil
}

// FindRequestCandidates returns policies that potentially match a ladon.Request
func (s *SQLManager) FindRequestCandidates(ctx context.Context, r *ladon.Request) (ladon.Policies, error) {
	rows, err := s.db.WithContext(ctx).Raw(s.stmts.QueryRequestCandidates, r.Subject, r.Subject).Rows()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func scanRows(rows *sql.Rows) (ladon.Policies, error) {
	var policies = map[string]*ladon.DefaultPolicy{}

	for rows.Next() {
		var p ladon.DefaultPolicy
		var conditions []byte
		var resource, subject, action sql.NullString
		p.Actions = []string{}
		p.Subjects = []string{}
		p.Resources = []string{}

		if err := rows.Scan(&p.ID, &p.Effect, &conditions, &p.Description, &p.Meta, &subject, &resource, &action); err == gorm.ErrRecordNotFound {
			return nil, ladon.NewErrResourceNotFound(err)
		} else if err != nil {
			return nil, errors.WithStack(err)
		}

		p.Conditions = ladon.Conditions{}
		if err := json.Unmarshal(conditions, &p.Conditions); err != nil {
			return nil, errors.WithStack(err)
		}

		if c, ok := policies[p.ID]; ok {
			if action.Valid {
				policies[p.ID].Actions = append(c.Actions, action.String)
			}

			if subject.Valid {
				policies[p.ID].Subjects = append(c.Subjects, subject.String)
			}

			if resource.Valid {
				policies[p.ID].Resources = append(c.Resources, resource.String)
			}
		} else {
			if action.Valid {
				p.Actions = []string{action.String}
			}

			if subject.Valid {
				p.Subjects = []string{subject.String}
			}

			if resource.Valid {
				p.Resources = []string{resource.String}
			}

			policies[p.ID] = &p
		}
	}

	var result = make(ladon.Policies, len(policies))
	var count int
	for _, v := range policies {
		v.Actions = uniq(v.Actions)
		v.Resources = uniq(v.Resources)
		v.Subjects = uniq(v.Subjects)
		result[count] = v
		count++
	}

	return result, nil
}

// GetAll returns all policies
func (s *SQLManager) GetAll(ctx context.Context, limit, offset int64) (ladon.Policies, error) {
	rows, err := s.db.WithContext(ctx).Raw(s.stmts.GetAllQuery, limit, offset).Rows()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// Get retrieves a policy.
func (s *SQLManager) Get(ctx context.Context, id string) (ladon.Policy, error) {
	rows, err := s.db.WithContext(ctx).Raw(s.stmts.GetQuery, id).Rows()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	policies, err := scanRows(rows)
	if err != nil {
		return nil, err
	} else if len(policies) == 0 {
		return nil, ladon.NewErrResourceNotFound(gorm.ErrRecordNotFound)
	}

	return policies[0], nil
}

// Delete removes a policy.
func (s *SQLManager) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.delete(id, tx)
	})
}

// Delete removes a policy.
func (s *SQLManager) delete(id string, tx *gorm.DB) error {
	if err := tx.Exec(s.stmts.DeletePolicy, id).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// FindPoliciesForSubject returns policies that could match the subject.
func (s *SQLManager) FindPoliciesForSubject(ctx context.Context, subject string) (ladon.Policies, error) {
	rows, err := s.db.WithContext(ctx).Raw(s.stmts.QueryRequestCandidates, subject, subject).Rows()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanRows(rows)
}

// FindPoliciesForResource returns policies that could match the resource.
func (s *SQLManager) FindPoliciesForResource(ctx context.Context, resource string) (ladon.Policies, error) {
	rows, err := s.db.WithContext(ctx).Raw(s.stmts.QueryRequestCandidates, resource, resource).Rows()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ladon.NewErrResourceNotFound(err)
		}
		return nil, errors.WithStack(err)
	}
	return scanRows(rows)
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
