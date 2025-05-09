#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

echo "Building control plane..."
cd "$WORKSPACE_ROOT/control-plane"
go build -o cp main.go

echo "Running control plane... (Press Ctrl+C to stop)"
./cp 
 