// Package models provides data models for the Ladon SQL Manager.
// This package contains GORM models for managing authorization policies,
// subjects, actions, and resources in a SQL database.
//
// The main entities are:
//   - Policy: Represents authorization policies with allow/deny effects
//   - Subject: Represents entities that can perform actions (users, roles, etc.)
//   - Action: Represents operations that can be performed
//   - Resource: Represents objects that actions can be performed on
//
// File Organization:
//   - constants.go: Contains all constants (table names, field sizes, effects)
//   - interfaces.go: Contains interfaces for validation and common operations
//   - base_types.go: Contains JSONText and BaseEntity types
//   - policy.go: Contains the Policy model and its methods
//   - entities.go: Contains Subject, Action, and Resource models
//   - relations.go: Contains relationship models (PolicySubjectRel, etc.)
//
// Example usage:
//
//	policy := &Policy{
//		ID:          "policy-1",
//		Description: "Allow users to read documents",
//		Effect:      EffectAllow,
//		Conditions:  JSONText(`{"user_role": "reader"}`),
//	}
//
//	if err := policy.Validate(); err != nil {
//		log.Fatal(err)
//	}
package models
