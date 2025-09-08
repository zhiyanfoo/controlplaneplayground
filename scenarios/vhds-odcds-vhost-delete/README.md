# VHDS-ODCDS Virtual Host Deletion Test

This test demonstrates VHDS (Virtual Host Discovery Service) with ODCDS (On-Demand Cluster Discovery Service) behavior when a referenced virtual host is deleted after successful requests.

## Test Scenario

This test verifies that when a virtual host is deleted from the VHDS configuration, Envoy properly:
1. Continues to serve existing requests using cached virtual host information
2. Eventually fails new requests when the virtual host is no longer available
3. Handles the virtual host deletion gracefully via VHDS
4. Returns appropriate 404 responses for requests to deleted virtual hosts

## Test Steps

You'll need to run these in separate tabs:

### 1. Start the control-plane
```
./scripts/build_and_run_control_plane.sh
```

### 2. Start Envoy
```
./scripts/run_envoy.sh
```

### 3. Start test backends
```
./scripts/start_test_services.sh
```

### 4. Load initial configuration with virtual host
```
go run cli/*.go -config scenarios/vhds-odcds-vhost-delete/initial-config.json -action update
```

### 5. Make initial requests (establish virtual host and cluster)
```
./scripts/make_request.sh
```
*Note: First request may fail due to xDS timing - this is expected*

```
./scripts/make_request.sh
```
*Should succeed - virtual host and cluster are established*

### 6. Delete the virtual host
```
go run cli/*.go -type-url "type.googleapis.com/envoy.config.route.v3.VirtualHost" -name "basic_grpc_route/localhost:10000" -action delete
```

### 7. Make requests after virtual host deletion
```
./scripts/make_request.sh
```
*Should fail with 404 - virtual host no longer available*

```
./scripts/make_request.sh
```
*Should continue to fail with 404*

## Expected Behavior

- Initial requests succeed and establish the virtual host via VHDS
- After virtual host deletion, new requests fail with 404 Not Found
- Envoy handles the virtual host deletion gracefully without crashing
- Error responses indicate the virtual host/route is no longer available
- Clusters may remain available but are unreachable without virtual host routes

## Key Test Points

1. **Virtual host deletion handling**: Envoy gracefully handles deleted virtual hosts
2. **404 response generation**: Appropriate 404 errors returned for missing virtual hosts
3. **VHDS behavior**: Discovery service properly removes deleted virtual hosts
4. **Route cleanup**: Routes are no longer available after virtual host deletion