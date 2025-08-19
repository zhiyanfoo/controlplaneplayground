# Docker Compose Setup (Recommended)

Start the complete environment with a single command:

```bash
docker compose up -d
```

This starts all services: control-plane, Envoy proxy, and test servers.

## Management Access

- **Control Plane Cache**: http://localhost:8734 (view xDS cache)
- **Envoy Admin Panel**: http://localhost:60001 (proxy stats, config dump)
- **CLI Management**: Build CLI locally and connect to control-plane

## Making Requests

First, configure xDS resources using the CLI:

```bash
# Build CLI locally
go build -o bin/cli ./cli

# Configure basic gRPC resources
./bin/cli -config base-resources/basic-grpc.json
```

Then test via Envoy proxy:

```bash
# Test gRPC request (port 10000 goes to Envoy)
./scripts/make_request.sh 10000

# Test HTTP request
./scripts/make_http_request.sh 10000
```

The request flow: **Client → Envoy (port 10000) → Backend Services**
- Envoy gets dynamic configuration from control-plane
- Control-plane serves xDS config on port 18000

## Stop Services

```bash
docker compose down
```

## Development Tips

```bash
# Rebuild images when code changes
docker compose up --build

# View logs
docker compose logs -f

# Rebuild specific service
docker compose build control-plane
```

## Debugging

Access the debug container with network troubleshooting tools:

```bash
# Shell into debug container
docker compose exec debug bash

# Inside the debug container, you can use:
# Test connectivity
ping control-plane
ping 172.20.0.10

# Test HTTP endpoints
curl http://control-plane:8734
curl http://envoy:60001

# Test gRPC services using server reflection
grpcurl -plaintext -d '{"name": "Debug User"}' test-server-grpc:50051 test.TestService/SayHello

# Test via Envoy proxy (requires xDS configuration)  
grpcurl -plaintext -d '{"name": "Via Envoy"}' envoy:10000 test.TestService/SayHello
```

---

# Manual Setup (Alternative)

You'll need to run multiple components in seperate shells

The control plane that serves the xds configuration
```
./scripts/build_and_run_control_plane.sh
```

Envoy instance
```
scripts/run_envoy.sh
```

Grpc test server
```
./scripts/build_and_run_test_server.sh
```

The make a grpc request with

```
./scripts/make_request.sh
```

The request will be made to Envoy, which should forward it to test-server.
Envoy should get it's configuration from the control-plane.


