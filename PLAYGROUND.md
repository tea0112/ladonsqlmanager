# 🎮 Ladon Policy Playground Guide

Welcome to the interactive playground for experimenting with the Ladon authorization framework! This guide will help you explore various policy scenarios and understand how authorization works.

## 🚀 Quick Start

### Prerequisites
- Go 1.16+ installed
- PostgreSQL database running
- Database schema created (run `data/postgresql.sql`)
- `DB_STRING` environment variable set

### Setup Your Environment
```bash
# Set your database connection
export DB_STRING="host=localhost user=root password='Root\!23456' dbname=tea sslmode=disable"

# Verify the main application works
go build
./ladonsqlmanager
```

## 🎯 Interactive Examples

### 1. **Basic Policy Creation and Testing**

Start with the example application to see basic policy management:

```bash
# Build and run the example
go build -o example/example example/main.go
./example/example
```

This will:
- Create a sample policy
- Test basic authorization
- Show how policies work

### 2. **Policy Scenarios to Try**

#### Scenario A: File System Access Control
```go
// Create a policy that allows users to read their own files
policy := &ladon.DefaultPolicy{
    ID:          "user-file-read",
    Description: "Users can read their own files",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"user"},
    Resources:   []string{"file:user:.*"},
    Actions:     []string{"read"},
}

// Test the policy
request := &ladon.Request{
    Subject:  "user",
    Resource: "file:user:document.txt",
    Action:   "read",
}

err := warden.IsAllowed(ctx, request)
// Result: ✅ Access ALLOWED
```

#### Scenario B: Role-Based Access Control
```go
// Admin policy - full access
adminPolicy := &ladon.DefaultPolicy{
    ID:          "admin-full",
    Description: "Admin has full access",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"admin"},
    Resources:   []string{".*"},
    Actions:     []string{".*"},
}

// User policy - limited access
userPolicy := &ladon.DefaultPolicy{
    ID:          "user-limited",
    Description: "Users have limited access",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"user"},
    Resources:   []string{"user:.*", "public:.*"},
    Actions:     []string{"read", "write"},
}

// Test scenarios
adminRequest := &ladon.Request{
    Subject:  "admin",
    Resource: "system:config",
    Action:   "delete",
}
// Result: ✅ Access ALLOWED

userRequest := &ladon.Request{
    Subject:  "user",
    Resource: "system:config",
    Action:   "delete",
}
// Result: ❌ Access DENIED
```

#### Scenario C: API Endpoint Protection
```go
// Public read access
publicPolicy := &ladon.DefaultPolicy{
    ID:          "api-public-read",
    Description: "Public read access to API",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"guest", "user", "admin"},
    Resources:   []string{"api:public:.*"},
    Actions:     []string{"GET"},
}

// User write access
userWritePolicy := &ladon.DefaultPolicy{
    ID:          "api-user-write",
    Description: "Users can write to user endpoints",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"user", "admin"},
    Resources:   []string{"api:user:.*"},
    Actions:     []string{"POST", "PUT", "DELETE"},
}

// Admin only access
adminOnlyPolicy := &ladon.DefaultPolicy{
    ID:          "api-admin-only",
    Description: "Admin only access to admin endpoints",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"admin"},
    Resources:   []string{"api:admin:.*"},
    Actions:     []string{".*"},
}
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
```go
// Tenant isolation
tenantPolicy := &ladon.DefaultPolicy{
    ID:          "tenant-isolation",
    Description: "Users can only access their tenant's data",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"user"},
    Resources:   []string{"tenant:<tenant_id>:.*"},
    Actions:     []string{"read", "write"},
}

// Test with different tenants
request1 := &ladon.Request{
    Subject:  "user@company1.com",
    Resource: "tenant:company1:users",
    Action:   "read",
}
// Result: ✅ Access ALLOWED

request2 := &ladon.Request{
    Subject:  "user@company1.com",
    Resource: "tenant:company2:users",
    Action:   "read",
}
// Result: ❌ Access DENIED
```

### Example 2: Department-Based Access
```go
// Engineering department access
engPolicy := &ladon.DefaultPolicy{
    ID:          "dept-engineering",
    Description: "Engineering department access",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"dept:engineering"},
    Resources:   []string{"dept:engineering:.*"},
    Actions:     []string{"read", "write"},
}

// Cross-department read access
crossDeptPolicy := &ladon.DefaultPolicy{
    ID:          "cross-dept-read",
    Description: "Read access across departments",
    Effect:      ladon.AllowAccess,
    Subjects:    []string{"dept:.*"},
    Resources:   []string{"dept:.*:public"},
    Actions:     []string{"read"},
}
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

## 🔍 Interactive Testing

### Using the Example Application
```bash
# Run the example
./example/example

# This will:
# 1. Create a sample policy
# 2. Test authorization
# 3. Show policy details
```

### Manual Testing with Go
```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/ladonsqlmanager"
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
    if err := manager.Init(); err != nil {
        log.Fatal(err)
    }
    
    // Create warden
    warden := &ladon.Ladon{Manager: manager}
    ctx := context.Background()
    
    // Test your policies here
    request := &ladon.Request{
        Subject:  "user",
        Resource: "file:user:document.txt",
        Action:   "read",
    }
    
    if err := warden.IsAllowed(ctx, request); err != nil {
        fmt.Printf("❌ Access DENIED: %v\n", err)
    } else {
        fmt.Println("✅ Access ALLOWED!")
    }
}
```

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

## 🚀 Next Steps

1. **Run the Example**: Start with `./example/example`
2. **Create Your Own Policies**: Experiment with different scenarios
3. **Test Authorization**: Use the warden to check permissions
4. **Build Real Applications**: Integrate Ladon into your projects

---

**🎮 Happy Policy Testing!** Use this playground to experiment with different authorization scenarios and build confidence in your Ladon implementation. The key is to start simple and gradually build complexity as you understand how the system works.
