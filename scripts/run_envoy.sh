#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

# Check if Envoy is installed
if ! command -v envoy &> /dev/null
then
    echo "Envoy command could not be found. Please install Envoy (e.g., 'brew install envoy') and ensure it's in your PATH."
    exit 1
fi

echo "Starting Envoy with config $WORKSPACE_ROOT/envoy/bootstrap.yaml... (Press Ctrl+C to stop)"
envoy -c "$WORKSPACE_ROOT/envoy/bootstrap.yaml" --log-level debug
