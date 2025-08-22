# �� How to Run the Ladon SQL Manager Project

## 📋 Prerequisites

Before running this project, you need:

1. **Go 1.16+** installed on your system
2. **PostgreSQL** or **MySQL** database server
3. **Git** for cloning the repository

## 🗄️ Database Setup

### Option 1: PostgreSQL (Recommended)

1. **Install PostgreSQL**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install postgresql postgresql-contrib
   
   # macOS with Homebrew
   brew install postgresql
   
   # Start PostgreSQL service
   sudo systemctl start postgresql  # Linux
   brew services start postgresql    # macOS
   ```

2. **Create Database and User**:
   ```bash
   sudo -u postgres psql
   
   CREATE DATABASE ladon_test;
   CREATE USER ladon_user WITH PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE ladon_test TO ladon_user;
   \q
   ```

3. **Run Schema Setup**:
   ```bash
   psql -h localhost -U ladon_user -d ladon_test -f data/postgresql.sql
   ```

### Option 2: MySQL

1. **Install MySQL**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install mysql-server
   
   # macOS with Homebrew
   brew install mysql
   
   # Start MySQL service
   sudo systemctl start mysql  # Linux
   brew services start mysql    # macOS
   ```

2. **Create Database and User**:
   ```bash
   sudo mysql -u root
   
   CREATE DATABASE ladon_test;
   CREATE USER 'ladon_user'@'localhost' IDENTIFIED BY 'your_password';
   GRANT ALL PRIVILEGES ON ladon_test.* TO 'ladon_user'@'localhost';
   FLUSH PRIVILEGES;
   EXIT;
   ```

3. **Run Schema Setup**:
   ```bash
   mysql -h localhost -u ladon_user -p ladon_test < data/mysql.sql
   ```

## 🔧 Environment Configuration

Set the database connection string as an environment variable:

```bash
# For PostgreSQL
export DB_STRING="host=localhost user=ladon_user password=your_password dbname=ladon_test sslmode=disable"

# For MySQL
export DB_STRING="ladon_user:your_password@tcp(localhost:3306)/ladon_test?charset=utf8mb4&parseTime=True&loc=Local"
```

## 🏃‍♂️ Running the Example Application

### 1. Build the Project

```bash
# Navigate to project directory
cd /home/vmo/workspace/go/ladonsqlmanager

# Build the main package
go build

# Build the example application
go build -o example/example example/main.go
```

### 2. Run the Example

```bash
# Make sure DB_STRING is set
echo $DB_STRING

# Run the example
./example/example
```

**Expected Output**:
```
2024/01/XX XX:XX:XX policy created successfully
2024/01/XX XX:XX:XX Found policies: [policy details]
2024/01/XX XX:XX:XX Authorization successful
```

## �� Using as a Library

### 1. Import the Package

```go
package main

import (
    "context"
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
    
    // Create warden for authorization
    warden := &ladon.Ladon{Manager: manager}
    
    // Use the manager...
    ctx := context.Background()
    policies, err := manager.GetAll(ctx, 10, 0)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d policies", len(policies))
}
```

### 2. Create Your Own Policies

```go
policy := &ladon.DefaultPolicy{
    ID:          "user-read-policy",
    Description: "Allow users to read their own data",
    Subjects:    []string{"user"},
    Effect:      ladon.AllowAccess,
    Resources:   []string{"user:<.*>"},
    Actions:     []string{"read"},
}

ctx := context.Background()
err := manager.Create(ctx, policy)
if err != nil {
    log.Fatal(err)
}
```

## 🧪 Testing the Setup

### 1. Verify Database Connection

```bash
# Test PostgreSQL connection
psql -h localhost -U ladon_user -d ladon_test -c "SELECT COUNT(*) FROM ladon_policy;"

# Test MySQL connection
mysql -h localhost -u ladon_user -p ladon_test -e "SELECT COUNT(*) FROM ladon_policy;"
```

### 2. Run Basic Tests

```bash
# Test if the package builds
go build

# Test if example builds
go build ./example

# Check for any linting issues
go vet ./...
```

## 🚨 Troubleshooting

### Common Issues and Solutions

1. **"Cannot connect to database"**
    - Verify database service is running
    - Check connection string format
    - Ensure firewall allows connections

2. **"Table doesn't exist"**
    - Run the schema setup SQL files
    - Check database name in connection string
    - Verify user has proper permissions

3. **"Import not found"**
    - Run `go mod tidy` to sync dependencies
    - Ensure you're in the correct directory
    - Check Go version compatibility

4. **"Permission denied"**
    - Verify database user permissions
    - Check if database exists
    - Ensure proper authentication

### Debug Mode

Enable verbose logging by setting:

```bash
export GORM_LOG_LEVEL=info
export GORM_LOG_SLOW_THRESHOLD=100ms
```

## 📖 Next Steps

After successfully running the example:

1. **Explore the API**: Try creating different types of policies
2. **Test Authorization**: Use the warden to check permissions
3. **Integrate**: Use the manager in your own Go applications
4. **Customize**: Modify the SQL statements for your specific needs

## �� Useful Resources

- [GORM Documentation](https://gorm.io/docs/)
- [Ladon Authorization Framework](https://github.com/ory/ladon)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [MySQL Documentation](https://dev.mysql.com/doc/)

---

**�� Congratulations!** You've successfully set up and run the Ladon SQL Manager project. The example demonstrates basic policy management and authorization, and you can now use this as a foundation for building more complex authorization systems in your Go applications.

```shellscript
   # Ubuntu/Debian
   sudo apt-get install postgresql postgresql-contrib
   
   # macOS with Homebrew
   brew install postgresql
   
   # Start PostgreSQL service
   sudo systemctl start postgresql  # Linux
   brew services start postgresql    # macOS
```

```shellscript
   sudo -u postgres psql
   
   CREATE DATABASE ladon_test;
   CREATE USER ladon_user WITH PASSWORD 'your_password';
   GRANT ALL PRIVILEGES ON DATABASE ladon_test TO ladon_user;
   \q
```

```shellscript
   psql -h localhost -U ladon_user -d ladon_test -f data/postgresql.sql
```

```shellscript
   # Ubuntu/Debian
   sudo apt-get install mysql-server
   
   # macOS with Homebrew
   brew install mysql
   
   # Start MySQL service
   sudo systemctl start mysql  # Linux
   brew services start mysql    # macOS
```

```shellscript
   sudo mysql -u root
   
   CREATE DATABASE ladon_test;
   CREATE USER 'ladon_user'@'localhost' IDENTIFIED BY 'your_password';
   GRANT ALL PRIVILEGES ON ladon_test.* TO 'ladon_user'@'localhost';
   FLUSH PRIVILEGES;
   EXIT;
```

```shellscript
   mysql -h localhost -u ladon_user -p ladon_test < data/mysql.sql
```

```shellscript
# For PostgreSQL
export DB_STRING="host=localhost user=ladon_user password=your_password dbname=ladon_test sslmode=disable"

# For MySQL
export DB_STRING="ladon_user:your_password@tcp(localhost:3306)/ladon_test?charset=utf8mb4&parseTime=True&loc=Local"
```

```shellscript
# Navigate to project directory
cd /home/vmo/workspace/go/ladonsqlmanager

# Build the main package
go build

# Build the example application
go build -o example/example example/main.go
```

```shellscript
# Make sure DB_STRING is set
echo $DB_STRING

# Run the example
./example/example
```

```plaintext
2024/01/XX XX:XX:XX policy created successfully
2024/01/XX XX:XX:XX Found policies: [policy details]
2024/01/XX XX:XX:XX Authorization successful
```

```go
package main

import (
    "context"
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
    
    // Create warden for authorization
    warden := &ladon.Ladon{Manager: manager}
    
    // Use the manager...
    ctx := context.Background()
    policies, err := manager.GetAll(ctx, 10, 0)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Found %d policies", len(policies))
}
```

```go
policy := &ladon.DefaultPolicy{
    ID:          "user-read-policy",
    Description: "Allow users to read their own data",
    Subjects:    []string{"user"},
    Effect:      ladon.AllowAccess,
    Resources:   []string{"user:<.*>"},
    Actions:     []string{"read"},
}

ctx := context.Background()
err := manager.Create(ctx, policy)
if err != nil {
    log.Fatal(err)
}
```

```shellscript
# Test PostgreSQL connection
psql -h localhost -U ladon_user -d ladon_test -c "SELECT COUNT(*) FROM ladon_policy;"

# Test MySQL connection
mysql -h localhost -u ladon_user -p ladon_test -e "SELECT COUNT(*) FROM ladon_policy;"
```

```shellscript
# Test if the package builds
go build

# Test if example builds
go build ./example

# Check for any linting issues
go vet ./...
```

```shellscript
export GORM_LOG_LEVEL=info
export GORM_LOG_SLOW_THRESHOLD=100ms
```

