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

echo "Sending request via grpcurl to Envoy (localhost:10000)..."

grpcurl -plaintext \
    -import-path "$WORKSPACE_ROOT/test-server" \
    -proto "$WORKSPACE_ROOT/test-server/test.proto" \
    -d '{"name": "Playground User"}' \
    -authority "localhost:10000" \
    localhost:10000 test.TestService/SayHello

if [ $? -eq 0 ]; then
    echo "grpcurl request successful!"
else
    echo "grpcurl request failed."
    exit 1 # Exit with error if grpcurl failed
fi 
