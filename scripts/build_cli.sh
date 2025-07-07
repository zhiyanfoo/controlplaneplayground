#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

echo "Building CLI tool..."
cd "$WORKSPACE_ROOT"

# Download dependencies
go mod tidy

# Build the CLI
go build -o cli/cli cli/main.go

echo "CLI tool built successfully as cli/cli"
echo ""
echo "Usage examples:"
echo "  ./cli/cli -config examples/update_resources.yaml -action update"
echo "  ./cli/cli -config examples/delete_resources.yaml -action delete"
echo "  ./cli/cli -config examples/update_resources.yaml -action update -server localhost:18000" 
