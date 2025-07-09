#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

# Default port
DEFAULT_PORT=10000

# Parse command line arguments
PORT=${1:-$DEFAULT_PORT}

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null
then
    echo "grpcurl command could not be found. Please install grpcurl (e.g., 'brew install grpcurl')."
    exit 1
fi

echo "Sending request via grpcurl to Envoy (localhost:$PORT)..."

# grpcurl -plaintext \
#     -d '{"name": "Playground User"}' \
#     -authority "localhost:$PORT" \
#     localhost:$PORT test.TestService/SayHello

grpcurl -plaintext \
  -import-path "$WORKSPACE_ROOT/testpb" \
  -proto "$WORKSPACE_ROOT/testpb/test.proto" \
  -d '{"name": "Test User"}' \
  "localhost:$PORT" test.TestService/SayHello

if [ $? -eq 0 ]; then
    echo "grpcurl request successful!"
else
    echo "grpcurl request failed."
    exit 1 # Exit with error if grpcurl failed
fi 
