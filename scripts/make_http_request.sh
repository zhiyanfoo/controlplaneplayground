#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

# Check if curl is installed
if ! command -v curl &> /dev/null
then
    echo "curl command could not be found. Please install curl."
    exit 1
fi

echo "Sending HTTP request via curl to Envoy HTTP/1.1 listener (localhost:10001)..."

# Make HTTP request to the test endpoint
# --fail makes curl fail on HTTP error responses (4xx, 5xx)
curl --fail \
    -X POST \
    -H "Content-Type: application/json" \
    -H "Host: localhost:10001" \
    -d '{"name": "Playground User"}' \
    http://localhost:10001/test/sayhello

if [ $? -eq 0 ]; then
    echo ""
    echo "curl request successful!"
else
    echo "curl request failed."
    exit 1
fi 
