# 🎮 Ladon Policy Playground

Welcome to the interactive playground for experimenting with the Ladon authorization framework! This playground provides various tools, examples, and scenarios to help you understand and test authorization policies using GORM-based database management.

## 🚀 Getting Started

### Prerequisites
- Go 1.16+ installed
- PostgreSQL database running (use `make up` to start with Docker)
- `DB_STRING` environment variable set or `config.env` file created

### Quick Setup

#### Option 1: Using Docker (Recommended)
```bash
# Start PostgreSQL database
make up

# Create config.env file
echo "DB_STRING=postgres://ladon_user:ladon_password@localhost:5432/ladon_db?sslmode=disable" > config.env

# Build and run playground tools
cd playground
make run-cli
```

#### Option 2: Manual Database Setup
```bash
# Set your database connection
export DB_STRING="postgres://ladon_user:ladon_password@localhost:5432/ladon_db?sslmode=disable"

# Or create config.env file
echo "DB_STRING=postgres://ladon_user:ladon_password@localhost:5432/ladon_db?sslmode=disable" > config.env

# Build the playground tools
cd playground
make build

# Run the interactive CLI
make run-cli
```

## 🎯 Interactive Policy CLI

The `policy_cli` tool provides an interactive menu-driven interface for:

1. **Create Policies** - Build custom authorization policies
2. **List Policies** - View all existing policies
3. **Test Authorization** - Test access requests against policies
4. **Sample Policies** - Create pre-built example policies
5. **Delete Policies** - Remove unwanted policies
6. **Policy Details** - Examine specific policy configurations

## 🧪 Sample Scenarios

### Scenario 1: File System Access Control
```bash
# Create a policy that allows users to read their own files
Policy ID: user-file-read
Description: Users can read their own files
Effect: allow
Subjects: user
Resources: file:user:<.*>
Actions: read

# Test the policy
Subject: user
Resource: file:user:document.txt
Action: read
# Result: ✅ Access ALLOWED
```

### Scenario 2: Role-Based Access Control
```bash
# Create admin policy
Policy ID: admin-full-access
Description: Admin has full access to everything
Effect: allow
Subjects: admin
Resources: .*
Actions: .*

# Create user policy
Policy ID: user-limited-access
Description: Users have limited access
Effect: allow
Subjects: user
Resources: user:.*,public:.*
Actions: read,write

# Test admin access
Subject: admin
Resource: system:config
Action: delete
# Result: ✅ Access ALLOWED

# Test user access
Subject: user
Resource: system:config
Action: delete
# Result: ❌ Access DENIED
```

### Scenario 3: API Endpoint Protection
```bash
# Create API access policies
Policy ID: api-public-read
Description: Public read access to API
Effect: allow
Subjects: guest,user,admin
Resources: api:public:.*
Actions: GET

Policy ID: api-user-write
Description: Users can write to user endpoints
Effect: allow
Subjects: user,admin
Resources: api:user:.*
Actions: POST,PUT
```

## 🛠️ Available Tools

### 1. Interactive Policy CLI (`policy_cli`)
- **Build**: `make cli`
- **Run**: `make run-cli`
- **Features**: Full interactive policy management

### 2. Quick Test Scenarios (`quick_test`)
- **Build**: `make quick-test`
- **Run**: `make run-quick-test`
- **Features**: Automated test scenarios and examples

### 3. Playground Launcher (`playground.sh`)
- **Run**: `./playground.sh`
- **Features**: Menu-driven tool launcher with automatic building

## 🔧 Configuration

### Environment Variables
The playground tools automatically read from `config.env` if available:

```bash
# config.env
DB_STRING=postgres://ladon_user:ladon_password@localhost:5432/ladon_db?sslmode=disable
```

### Database Schema
The database schema is automatically created when you run any tool. GORM handles:
- Table creation
- Indexes and constraints
- Foreign key relationships
- Automatic migrations

## 📚 Usage Examples

### Running Individual Tools
```bash
# Build and run CLI
cd playground
make run-cli

# Build and run quick test
make run-quick-test

# Build all tools
make build
```

### Using the Playground Launcher
```bash
# Start the interactive launcher
cd playground
./playground.sh

# Choose from the menu:
# 1. Interactive Policy CLI
# 2. Quick Test Scenarios
# 3. Build All Tools
# 4. Show Database Status
# 5. Run Sample Policies
# 6. Run Database Migrations
```

### Database Management
```bash
# From the root directory
make up          # Start PostgreSQL
make migrate     # Run migrations
make status      # Check database status
make down        # Stop PostgreSQL
```

## 🧹 Cleanup

### Remove Build Artifacts
```bash
cd playground
make clean
```

### Reset Database
```bash
# From root directory
make reset-db
```

## 🔍 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Ensure PostgreSQL is running: `make up`
   - Check `config.env` or `DB_STRING` environment variable
   - Verify database credentials

2. **Schema Creation Failed**
   - Run migrations manually: `make migrate`
   - Check database permissions
   - Ensure database exists

3. **Build Failures**
   - Ensure Go 1.16+ is installed
   - Check import paths in Go files
   - Run `go mod tidy` if needed

### Getting Help
- Check the main project README for database setup
- Use `make help` for available commands
- Run `./playground.sh` for interactive help

## 🎉 Happy Policy Testing!

The playground is designed to make authorization policy testing fun and educational. Experiment with different scenarios, test edge cases, and build complex policies to understand how Ladon works!
