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

# Add basic gRPC listener and cluster
./cli add-listener base-resources/basic-grpc.json
./cli add-cluster base-resources/basic-grpc.json
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


