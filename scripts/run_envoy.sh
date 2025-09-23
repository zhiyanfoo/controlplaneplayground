#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

# Use ENVOY environment variable if set, otherwise use default
if [[ -n "$ENVOY" ]]; then
  ENVOY_PATH="$ENVOY"
else
  ENVOY_PATH="/opt/homebrew/bin/envoy"
fi

# Check if Envoy binary exists
if [[ ! -f "$ENVOY_PATH" ]]; then
    echo "Error: Envoy binary not found at $ENVOY_PATH"
    echo "Either install Envoy at the default location or set ENVOY environment variable to the correct path"
    echo "Example: ENVOY=/path/to/envoy $0"
    exit 1
fi

echo "Starting Envoy with config $WORKSPACE_ROOT/envoy/bootstrap.yaml... (Press Ctrl+C to stop)"
echo "Using Envoy binary: $ENVOY_PATH"
$ENVOY_PATH -c "$WORKSPACE_ROOT/envoy/bootstrap.yaml" \
  --log-level debug \
  --concurrency 1
