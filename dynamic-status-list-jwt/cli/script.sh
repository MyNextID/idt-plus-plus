#!/bin/bash

# Script to test the basic capabilities

# Check if Go is installed
if ! command -v go &> /dev/null
then
    echo "Go is not installed. Please install Go first."
    exit 1
fi

# Clone repository and navigate to CLI directory
clone_repo() {
    git clone git@github.com:MyNextID/idt-plus-plus.git
    cd idt-plus-plus/dynamic-status-list-jwt/cli || exit
}

# Install dependencies and build the CLI
install_dependencies() {
    go mod tidy
    go build
}

# Verify the installation by showing the help command
verify_installation() {
    ./dsl --help
}

# Issue a mock JWT
issue_mock_jwt() {
    ./dsl issue
}

# Create a new status list entry
create_status_list_entry() {
    local input_file="${1:-mock-jwt.json}" # Default to 'mock-jwt.json' if no input is provided
    ./dsl new -i "$input_file"
}

# Compute the revocation identifier
compute_revocation_identifier() {
    local input_file="${1:-mock-jwt.json}" # Default to 'mock-jwt.json' if no input is provided
    local timestamp="${2:-$(date +%s)}"    # Default to current timestamp if no timestamp is provided
    ./dsl wallet -i "$input_file" -t "$timestamp"
}

# Recompute the dynamic status list
recompute_dsl() {
    local timestamp="${1:-$(date +%s)}"    # Default to current timestamp if no timestamp is provided
    ./dsl recompute -t "$timestamp"
}

# Revoke a JWT using its jti identifier
revoke_jwt() {
    local jti="${1:-123}" # Default to '123' if no jti is provided (example from the README)
    ./dsl revoke --jti "$jti"
}

# Verify JWT revocation status
verify_revocation_status() {
    local status_list_file="${1:-dsl.json}"  # Default to 'dsl.json' if no input is provided
    local holder_file="${2:-holder_status-list-identifier.json}" # Default to 'holder_status-list-identifier.json'
    ./dsl verify -s "$status_list_file" -p "$holder_file"
}

# Create detached status list metadata
create_detached_metadata() {
    local input_file="${1:-mock-jwt.json}" # Default to 'mock-jwt.json' if no input is provided
    ./dsl new -i "$input_file" --detached
}

# Main menu function to provide CLI commands to the user
menu() {
    echo "Dynamic Status List CLI"
    echo "1. Clone the repository and build"
    echo "2. Issue a mock JWT"
    echo "3. Create a new status list entry"
    echo "4. Compute the revocation identifier"
    echo "5. Recompute the dynamic status list"
    echo "6. Revoke a JWT"
    echo "7. Verify JWT revocation status"
    echo "8. Create detached status list metadata"
    echo "9. Exit"
    
    read -p "Choose an option: " option
    
    case $option in
        1)
            clone_repo
            install_dependencies
            verify_installation
            ;;
        2)
            issue_mock_jwt
            ;;
        3)
            read -p "Enter input file for status list entry (default: mock-jwt.json): " input_file
            create_status_list_entry "${input_file:-mock-jwt.json}"
            ;;
        4)
            read -p "Enter input file for revocation identifier (default: mock-jwt.json): " input_file
            read -p "Enter timestamp (leave empty for current time): " timestamp
            compute_revocation_identifier "${input_file:-mock-jwt.json}" "${timestamp:-$(date +%s)}"
            ;;
        5)
            read -p "Enter timestamp (leave empty for current time): " timestamp
            recompute_dsl "${timestamp:-$(date +%s)}"
            ;;
        6)
            read -p "Enter jti to revoke (default: 123): " jti
            revoke_jwt "${jti:-123}"
            ;;
        7)
            read -p "Enter status list file (default: dsl.json): " status_list_file
            read -p "Enter holder file (default: holder_status-list-identifier.json): " holder_file
            verify_revocation_status "${status_list_file:-dsl.json}" "${holder_file:-holder_status-list-identifier.json}"
            ;;
        8)
            read -p "Enter input file for detached metadata (default: mock-jwt.json): " input_file
            create_detached_metadata "${input_file:-mock-jwt.json}"
            ;;
        9)
            exit 0
            ;;
        *)
            echo "Invalid option. Please choose again."
            menu
            ;;
    esac
}

# Run the menu function
menu
