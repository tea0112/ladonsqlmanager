# Models Package

This package contains GORM models for the Ladon SQL Manager. It provides
schema definitions and helpers for managing authorization policies, subjects,
actions, and resources stored in a SQL database.

## Contents

- constants.go — constants for effects, table names, and field limits
- interfaces.go — common interfaces (Validator, Entity, PolicyEntity)
- base_types.go — JSONText and BaseEntity shared types
- policy.go — Policy model and helpers
- entities.go — Subject, Action, Resource models
- relations.go — Relationship tables between Policy and entities
- models.go — package documentation and overview

## Design goals

- Separation of concerns: small, focused files per concept
- Strong typing and validation helpers
- Testable via small interfaces
- Consistent naming and table mapping via constants

## Data model overview

- Policy: top-level rule with effect (allow/deny), conditions (JSON), and metadata
- Subject: who performs actions (e.g., user, role)
- Action: what is attempted (e.g., read, write)
- Resource: what is acted upon (e.g., document)
- Relations: many-to-many associations between Policy and Subject/Action/Resource

## Key types

### JSONText

A lightweight wrapper for JSON payloads stored as text in the DB. Implements
sql.Scanner, driver.Valuer, and JSON marshal/unmarshal. Includes helpers:

- IsValid() — true if valid JSON or nil
- IsNull() — true if nil or literal null
- String() — string form for logging/debugging

Example:

```go
conditions := models.JSONText(`{"user_role":"reader"}`)
if !conditions.IsValid() { /* handle invalid JSON */ }
```

### BaseEntity

Shared fields and validation logic for Subject, Action, and Resource:

- ID (varchar(64), primary key)
- HasRegex (bool)
- Compiled (varchar(511), unique)
- Template (varchar(511), unique)
- CreatedAt / UpdatedAt
- DeletedAt (soft delete)

All entity types embed BaseEntity and inherit its Validate method.

### Interfaces

- Validator: Validate() error
- Entity: TableName() string + Validate()
- PolicyEntity: GetID() string + Entity

These facilitate testing and mocking.

## Constants

- Effects: EffectAllow, EffectDeny
- Table names: TableNamePolicy, TableNameSubject, TableNameAction,
  TableNameResource, TableNamePolicySubjectRel, TableNamePolicyActionRel,
  TableNamePolicyResourceRel
- Field limits: PolicyIDMaxLength, EntityIDMaxLength, CompiledMaxLength,
  TemplateMaxLength

## GORM mapping

- Soft deletes via gorm.DeletedAt
- Explicit table names via TableName() using constants
- Many-to-many relations defined on Policy and via explicit relation structs

## Usage examples

### Define and validate a policy

```go
p := &models.Policy{
    ID:          "policy-1",
    Description: "Allow readers to view documents",
    Effect:      models.EffectAllow,
    Conditions:  models.JSONText(`{"user_role":"reader"}`),
}

if err := p.Validate(); err != nil {
    return err
}
```

### Attach entities to a policy (GORM)

```go
// Assuming db is *gorm.DB and you have subjects/actions/resources created
if err := db.Create(p).Error; err != nil { return err }

if err := db.Model(p).Association("Subjects").Append(&subject).Error; err != nil { /* ... */ }
if err := db.Model(p).Association("Actions").Append(&action).Error; err != nil { /* ... */ }
if err := db.Model(p).Association("Resources").Append(&resource).Error; err != nil { /* ... */ }
```

### Auto-migrate schema

```go
if err := db.AutoMigrate(
    &models.Policy{},
    &models.Subject{},
    &models.Action{},
    &models.Resource{},
    &models.PolicySubjectRel{},
    &models.PolicyActionRel{},
    &models.PolicyResourceRel{},
); err != nil {
    log.Fatal(err)
}
```

### Validate entities

```go
s := &models.Subject{ BaseEntity: models.BaseEntity{ID: "role:reader", Template: "reader", Compiled: "^reader$"} }
if err := s.Validate(); err != nil { /* handle */ }
```

## Conventions and notes

- All table names are constants; change once if you rename tables
- Field length limits are centralized in constants
- JSONText enforces valid JSON on Unmarshal and Scan; use IsValid before persisting literals
- Soft deletes are enabled on all core entities by default

## Testing tips

- Prefer testing via the Validator interface
- For DB tests, use a transaction per test and rollback at the end
- Use the constants for table names in raw queries

## Maintenance

When adding a new entity type:

1. Consider embedding BaseEntity
2. Add Validate method or reuse BaseEntity.Validate
3. Add TableName() using a new constant in constants.go
4. Update AutoMigrate in your initialization code

## Build

```bash
go build ./...
```

