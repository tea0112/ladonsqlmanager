#!/bin/bash

# 🎮 Ladon Policy Playground Launcher
# This script helps you run various playground tools and examples

set -e

echo "🎮 Welcome to Ladon Policy Playground!"
echo "====================================="

# Check if config.env exists and load it
if [ -f "config.env" ]; then
    echo "📁 Found config.env, loading configuration..."
    export $(cat config.env | grep -v '^#' | xargs)
    echo "✅ Configuration loaded from config.env"
else
    echo "⚠️  No config.env found, checking for DB_STRING environment variable..."
fi

# Check if DB_STRING is set
if [ -z "$DB_STRING" ]; then
    echo "❌ Error: DB_STRING environment variable not set"
    echo "Please either:"
    echo "  1. Create a config.env file with your database configuration, or"
    echo "  2. Set the DB_STRING environment variable"
    echo ""
    echo "Example config.env:"
    echo "  DB_STRING=postgres://ladon_user:ladon_password@localhost:5432/ladon_db?sslmode=disable"
    exit 1
fi

echo "✅ Database connection configured"
echo ""

# Function to build tools
build_tools() {
    echo "🔨 Building playground tools..."
    
    # Build CLI tool
    echo "Building policy CLI..."
    cd playground
    go build -o policy_cli policy_cli.go
    cd ..
    
    # Build quick test tool
    echo "Building quick test tool..."
    cd playground
    go build -o quick_test quick_test_main.go
    cd ..
    
    echo "✅ Tools built successfully!"
    echo ""
}

# Function to show menu
show_menu() {
    echo "📋 Available Playground Tools:"
    echo "1. Interactive Policy CLI"
    echo "2. Quick Test Scenarios"
    echo "3. Build All Tools"
    echo "4. Show Database Status"
    echo "5. Run Sample Policies"
    echo "6. Run Database Migrations"
    echo "0. Exit"
    echo ""
}

# Function to run sample policies
run_sample_policies() {
    echo "🎯 Running sample policies..."
    cd playground
    if [ ! -f "policy_cli" ]; then
        echo "Policy CLI not built. Building first..."
        go build -o policy_cli policy_cli.go
    fi
    echo "4" | ./policy_cli
    cd ..
    echo ""
}

# Function to check database status
check_db_status() {
    echo "🔍 Checking database status..."
    
    # Extract database info from DB_STRING
    if [[ $DB_STRING == *"dbname=ladon_db"* ]]; then
        echo "Database: ladon_db"
    fi
    
    if [[ $DB_STRING == *"user=ladon_user"* ]]; then
        echo "User: ladon_user"
    fi
    
    echo "Connection: $DB_STRING"
    echo ""
}

# Function to run migrations
run_migrations() {
    echo "🗄️  Running database migrations..."
    cd ..
    go run cmd/migrate/main.go -action=migrate
    cd playground
    echo "✅ Migrations completed!"
    echo ""
}

# Main loop
while true; do
    show_menu
    read -p "Enter your choice: " choice
    
    case $choice in
        1)
            echo "🚀 Starting Interactive Policy CLI..."
            cd playground
            if [ ! -f "policy_cli" ]; then
                echo "Tool not built. Building first..."
                go build -o policy_cli policy_cli.go
            fi
            ./policy_cli
            cd ..
            ;;
        2)
            echo "🧪 Running Quick Test Scenarios..."
            cd playground
            if [ ! -f "quick_test" ]; then
                echo "Tool not built. Building first..."
                go build -o quick_test quick_test_main.go
            fi
            ./quick_test
            cd ..
            ;;
        3)
            build_tools
            ;;
        4)
            check_db_status
            ;;
        5)
            run_sample_policies
            ;;
        6)
            run_migrations
            ;;
        0)
            echo "👋 Goodbye! Happy policy testing!"
            exit 0
            ;;
        *)
            echo "❌ Invalid choice. Please try again."
            ;;
    esac
    
    echo ""
    read -p "Press Enter to continue..."
    clear
done
