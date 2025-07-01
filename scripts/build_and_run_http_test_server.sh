#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

echo "Building HTTP test server..."
cd "$WORKSPACE_ROOT/test-server-http"
go build -o http_server main.go

echo "Running HTTP test server... (Press Ctrl+C to stop)"
./http_server 
