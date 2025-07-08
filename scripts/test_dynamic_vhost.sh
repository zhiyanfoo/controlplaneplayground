#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null
then
    echo "grpcurl command could not be found. Please install grpcurl (e.g., 'brew install grpcurl')."
    exit 1
fi

echo "Testing dynamic virtual host on localhost:10002..."

grpcurl -plaintext \
    -import-path "$WORKSPACE_ROOT/test-server" \
    -proto "$WORKSPACE_ROOT/test-server/test.proto" \
    -d '{"name": "Dynamic Test User"}' \
    -authority "localhost:10002" \
    localhost:10002 test.TestService/SayHello

if [ $? -eq 0 ]; then
    echo "Dynamic virtual host test successful!"
else
    echo "Dynamic virtual host test failed."
    exit 1
fi 
