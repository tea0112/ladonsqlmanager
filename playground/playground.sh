#!/bin/bash

# 🎮 Ladon Policy Playground Launcher
# This script helps you run various playground tools and examples

set -e

echo "🎮 Welcome to Ladon Policy Playground!"
echo "====================================="

# Check if DB_STRING is set
if [ -z "$DB_STRING" ]; then
    echo "❌ Error: DB_STRING environment variable not set"
    echo "Please set it first:"
    echo "  export DB_STRING=\"host=localhost user=root password='Root\!23456' dbname=tea sslmode=disable\""
    exit 1
fi

echo "✅ Database connection configured"
echo ""

# Function to build tools
build_tools() {
    echo "🔨 Building playground tools..."
    
    # Build CLI tool
    echo "Building policy CLI..."
    cd playground/cli
    go build -o policy_cli .
    cd ../..
    
    # Build quick test tool
    echo "Building quick test tool..."
    cd playground/quick_test
    go build -o quick_test .
    cd ../..
    
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
    echo "0. Exit"
    echo ""
}

# Function to run sample policies
run_sample_policies() {
    echo "🎯 Running sample policies..."
    cd playground/cli
    echo "4" | ./policy_cli
    cd ../..
    echo ""
}

# Function to check database status
check_db_status() {
    echo "🔍 Checking database status..."
    
    # Extract database info from DB_STRING
    if [[ $DB_STRING == *"dbname=tea"* ]]; then
        echo "Database: tea"
    fi
    
    if [[ $DB_STRING == *"user=root"* ]]; then
        echo "User: root"
    fi
    
    echo "Connection: $DB_STRING"
    echo ""
}

# Main loop
while true; do
    show_menu
    read -p "Enter your choice: " choice
    
    case $choice in
        1)
            echo "🚀 Starting Interactive Policy CLI..."
            if [ ! -f "playground/cli/policy_cli" ]; then
                echo "Tool not built. Building first..."
                build_tools
            fi
            cd playground/cli
            ./policy_cli
            cd ../..
            ;;
        2)
            echo "🧪 Running Quick Test Scenarios..."
            if [ ! -f "playground/quick_test/quick_test" ]; then
                echo "Tool not built. Building first..."
                build_tools
            fi
            cd playground/quick_test
            ./quick_test
            cd ../..
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
