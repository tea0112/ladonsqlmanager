# 🎮 Ladon Policy Playground

Welcome to the interactive playground for experimenting with the Ladon authorization framework! This playground provides various tools, examples, and scenarios to help you understand and test authorization policies.

## 🚀 Getting Started

### Prerequisites
- Go 1.16+ installed
- PostgreSQL database running
- Database schema created (run `data/postgresql.sql`)
- `DB_STRING` environment variable set

### Quick Setup
```bash
# Set your database connection
export DB_STRING="host=localhost user=root password='Root\!23456' dbname=tea sslmode=disable"

# Build the playground tools
go build -o playground/policy_cli playground/policy_cli.go

# Run the interactive CLI
./playground/policy_cli
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
Actions: POST,PUT,DELETE

Policy ID: api-admin-only
Description: Admin only access to admin endpoints
Effect: allow
Subjects: admin
Resources: api:admin:.*
Actions: .*

# Test various scenarios
Subject: guest, Resource: api:public:users, Action: GET
# Result: ✅ Access ALLOWED

Subject: user, Resource: api:user:profile, Action: PUT
# Result: ✅ Access ALLOWED

Subject: user, Resource: api:admin:users, Action: DELETE
# Result: ❌ Access DENIED
```

## 🔐 Policy Patterns

### 1. **Wildcard Patterns**
- `.*` - Matches anything
- `user:.*` - Matches anything starting with "user:"
- `.*:read` - Matches anything ending with ":read"

### 2. **Specific Patterns**
- `admin` - Matches exactly "admin"
- `file:user:123` - Matches exactly this resource
- `read` - Matches exactly this action

### 3. **Complex Patterns**
- `file:user:<.*>` - Uses Ladon's template syntax
- `api:.*:read` - Matches any API endpoint with read action
- `user:.*:profile` - Matches user profile resources

## 🎭 Real-World Examples

### Example 1: Multi-Tenant SaaS Application
```bash
# Tenant isolation
Policy ID: tenant-isolation
Description: Users can only access their tenant's data
Effect: allow
Subjects: user
Resources: tenant:<tenant_id>:.*
Actions: read,write

# Test with different tenants
Subject: user@company1.com
Resource: tenant:company1:users
Action: read
# Result: ✅ Access ALLOWED

Subject: user@company1.com
Resource: tenant:company2:users
Action: read
# Result: ❌ Access DENIED
```

### Example 2: Time-Based Access Control
```bash
# Business hours access
Policy ID: business-hours
Description: Access only during business hours
Effect: allow
Subjects: employee
Resources: office:.*
Actions: access
Conditions: {"time": {"after": "09:00", "before": "17:00"}}

# Test during business hours
Subject: employee
Resource: office:main
Action: access
# Result: ✅ Access ALLOWED (if within time window)
```

### Example 3: Resource Hierarchy
```bash
# Department-based access
Policy ID: dept-access
Description: Department members can access their resources
Effect: allow
Subjects: dept:engineering
Resources: dept:engineering:.*
Actions: read,write

Policy ID: cross-dept-read
Description: Read access across departments
Effect: allow
Subjects: dept:.*
Resources: dept:.*:public
Actions: read

# Test department access
Subject: dept:engineering
Resource: dept:engineering:code
Action: write
# Result: ✅ Access ALLOWED

Subject: dept:marketing
Resource: dept:engineering:code
Action: read
# Result: ❌ Access DENIED
```

## 🧪 Testing Strategies

### 1. **Positive Testing**
- Test scenarios that should be allowed
- Verify policies work as expected
- Check edge cases within allowed ranges

### 2. **Negative Testing**
- Test scenarios that should be denied
- Verify security boundaries
- Check unexpected access patterns

### 3. **Boundary Testing**
- Test at the edges of policy definitions
- Verify regex patterns work correctly
- Check template syntax boundaries

### 4. **Performance Testing**
- Test with large numbers of policies
- Verify query performance
- Check memory usage patterns

## 🔍 Debugging Tips

### 1. **Policy Matching**
- Use the "Show Policy Details" feature to examine policies
- Check regex patterns for syntax errors
- Verify template delimiters

### 2. **Database Issues**
- Check database connection
- Verify schema is created correctly
- Check table permissions

### 3. **Policy Logic**
- Understand Ladon's evaluation order
- Check for conflicting policies
- Verify effect (allow/deny) settings

## 📚 Learning Path

### Beginner Level
1. **Start Simple**: Create basic allow/deny policies
2. **Test Basic Scenarios**: Simple subject/resource/action combinations
3. **Understand Patterns**: Learn regex and template syntax

### Intermediate Level
1. **Complex Policies**: Multi-subject, multi-resource policies
2. **Policy Relationships**: Understand how policies interact
3. **Real-world Scenarios**: Model actual application requirements

### Advanced Level
1. **Conditional Policies**: Time, location, and context-based access
2. **Performance Optimization**: Efficient policy design
3. **Custom Extensions**: Extend Ladon for specific needs

## 🚨 Common Pitfalls

### 1. **Regex Complexity**
- Overly complex patterns can be hard to debug
- Test regex patterns separately
- Use simple patterns when possible

### 2. **Policy Conflicts**
- Multiple policies can create unexpected results
- Understand Ladon's evaluation order
- Test policy combinations thoroughly

### 3. **Resource Naming**
- Inconsistent naming can cause access issues
- Use clear, hierarchical naming conventions
- Document resource naming patterns

## 🎉 Advanced Features

### 1. **Custom Conditions**
- Extend policies with custom logic
- Implement time-based access
- Add location-based restrictions

### 2. **Policy Inheritance**
- Create base policies and extend them
- Implement role hierarchies
- Build policy templates

### 3. **Audit and Logging**
- Track policy decisions
- Monitor access patterns
- Generate compliance reports

## 🔗 Useful Resources

- [Ladon Framework Documentation](https://github.com/ory/ladon)
- [GORM Documentation](https://gorm.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Regular Expressions Guide](https://regex101.com/)

---

**🎮 Happy Policy Testing!** Use this playground to experiment with different authorization scenarios and build confidence in your Ladon implementation.
