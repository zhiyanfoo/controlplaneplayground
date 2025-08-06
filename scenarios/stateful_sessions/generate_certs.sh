#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
CERTS_DIR="$SCRIPT_DIR/certs"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating self-signed TLS certificates for Envoy...${NC}"

# Create certs directory if it doesn't exist
mkdir -p "$CERTS_DIR"

# Generate private key and certificate
openssl req -x509 -newkey rsa:4096 -keyout "$CERTS_DIR/key.pem" -out "$CERTS_DIR/cert.pem" -days 365 -nodes -subj "/CN=localhost"

echo -e "${GREEN}✓ Generated certs/cert.pem and certs/key.pem${NC}"
echo -e "${YELLOW}Certificate details:${NC}"
openssl x509 -in "$CERTS_DIR/cert.pem" -text -noout | grep -E "(Subject:|Not Before:|Not After:)"