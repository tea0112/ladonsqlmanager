# Ladon SQL Manager

A GORM-based implementation of the Ladon policy manager for Go applications. This package provides a clean, modern interface for managing Ladon policies using GORM ORM with automatic migrations.

## Features

- **GORM Integration**: Full GORM ORM support with automatic migrations
- **Database Agnostic**: Supports both MySQL and PostgreSQL
- **Type Safety**: Strongly typed models with proper relationships
- **Automatic Schema Management**: Database schema is automatically created and updated via GORM
- **Ladon Compatible**: Implements the standard Ladon Manager interface
- **No Raw SQL**: Everything is handled through GORM models and migrations

## Installation

```bash
go get github.com/yourusername/ladonsqlmanager
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/yourusername/ladonsqlmanager"
    "github.com/ory/ladon"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // Connect to database
    db, err := gorm.Open(postgres.Open("your_connection_string"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Create manager
    manager := ladonsqlmanager.New(db, "postgres")
    
    // Initialize database schema (GORM will create all tables automatically)
    if err := manager.Init(); err != nil {
        log.Fatal(err)
    }

    // Create a policy
    policy := &ladon.DefaultPolicy{
        ID:          "policy-1",
        Description: "Allow users to read articles",
        Subjects:    []string{"user"},
        Effect:      ladon.AllowAccess,
        Resources:   []string{"article:*"},
        Actions:     []string{"read"},
    }

    ctx := context.Background()
    if err := manager.Create(ctx, policy); err != nil {
        log.Fatal(err)
    }

    // Check if access is allowed
    request := &ladon.Request{
        Subject:  "user",
        Resource: "article:123",
        Action:   "read",
    }

    policies, err := manager.FindRequestCandidates(ctx, request)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d matching policies", len(policies))
}
```

## Database Support

### PostgreSQL
```go
import "gorm.io/driver/postgres"

db, err := gorm.Open(postgres.Open("postgres://user:pass@localhost/dbname"), &gorm.Config{})
manager := ladonsqlmanager.New(db, "postgres")
```

### MySQL
```go
import "gorm.io/driver/mysql"

db, err := gorm.Open(mysql.Open("user:pass@tcp(localhost:3306)/dbname"), &gorm.Config{})
manager := ladonsqlmanager.New(db, "mysql")
```

## Models

The package provides the following GORM models:

- **Policy**: Main policy table with conditions and metadata
- **Subject**: Subject definitions with regex support
- **Action**: Action definitions with regex support
- **Resource**: Resource definitions with regex support
- **PolicySubjectRel**: Many-to-many relationship between policies and subjects
- **PolicyActionRel**: Many-to-many relationship between policies and actions
- **PolicyResourceRel**: Many-to-many relationship between policies and resources

## Database Schema

The schema is automatically created when you call `manager.Init()`. GORM handles:

- All necessary tables with proper foreign key constraints
- Indexes for performance optimization
- Support for soft deletes
- Automatic timestamp management
- Database-specific optimizations

## Migration Functions

```go
// Run migrations
err := ladonsqlmanager.Migrate(db)

// Drop all tables (useful for testing)
err := ladonsqlmanager.DropTables(db)

// Reset database (drop and recreate)
err := ladonsqlmanager.ResetDatabase(db)
```

## Command Line Migration Tool

The package includes a standalone migration tool:

```bash
# Run migrations
go run cmd/migrate/main.go -action=migrate -db="your_connection_string"

# Drop tables
go run cmd/migrate/main.go -action=drop -db="your_connection_string"

# Reset database
go run cmd/migrate/main.go -action=reset -db="your_connection_string"
```

## Docker Development Environment

Use the included Docker Compose setup for easy development:

```bash
# Start PostgreSQL and pgAdmin
make up

# Run migrations
make migrate

# Check status
make status

# Stop services
make down
```

## API Compatibility

This package maintains full compatibility with the Ladon Manager interface:

- `Create(ctx, policy)`: Create a new policy
- `Update(ctx, policy)`: Update an existing policy
- `Delete(ctx, id)`: Delete a policy
- `Get(ctx, id)`: Retrieve a policy by ID
- `GetAll(ctx, limit, offset)`: Get all policies with pagination
- `FindRequestCandidates(ctx, request)`: Find policies matching a request
- `FindPoliciesForSubject(ctx, subject)`: Find policies for a subject
- `FindPoliciesForResource(ctx, resource)`: Find policies for a resource

## Key Benefits

- **No Raw SQL**: All database operations use GORM
- **Automatic migrations**: Database schema is managed automatically
- **Type safety**: Strongly typed models instead of raw SQL queries
- **Better performance**: GORM's query optimization and connection pooling
- **Easier testing**: Mock database connections for unit tests
- **Database agnostic**: Same code works with MySQL and PostgreSQL

## Testing

```go
// Use an in-memory SQLite database for testing
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
if err != nil {
    t.Fatal(err)
}

manager := ladonsqlmanager.New(db, "sqlite")
if err := manager.Init(); err != nil {
    t.Fatal(err)
}

// Your test code here
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.