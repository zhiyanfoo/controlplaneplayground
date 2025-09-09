# VHDS-ODCDS Cluster Deletion Test

This test demonstrates VHDS (Virtual Host Discovery Service) with ODCDS (On-Demand Cluster Discovery Service) behavior when a referenced cluster is deleted after successful requests.

## Test Scenario

This test verifies that when a cluster referenced by a virtual host route is deleted, Envoy properly:
1. Continues to serve existing requests using cached cluster information
2. Eventually fails new requests when the cluster is no longer available
3. Handles the cluster deletion gracefully via ODCDS
4. Returns appropriate error responses for requests to deleted clusters

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

### 4. Load initial configuration with cluster
```
go run cli/*.go -config scenarios/vhds-odcds-cluster-delete/initial-config.json -action update
```

### 5. Start persistent client to observe connection behavior during cluster deletion
```
go run pinger/*.go
```
*Note: First request may fail due to xDS timing - this is expected*

*Let this run in the background to observe how the persistent connection behaves when the cluster is deleted*

### 6. Wait for successful requests, then delete the cluster (while pinger is running)
```
go run cli/*.go -type-url "type.googleapis.com/envoy.config.cluster.v3.Cluster" -name "basic_grpc_cluster" -action delete
```

*Observe in the pinger output how requests transition from success to failure, while the gRPC connection itself remains active*

## Expected Behavior

- Initial requests succeed and establish the cluster via ODCDS
- After cluster deletion, new requests fail with appropriate error responses
- Envoy handles the cluster deletion gracefully without crashing
- Error responses indicate the cluster is no longer available
- No service interruption for other unaffected clusters (if any)

## Key Test Points

1. **Cluster deletion handling**: Envoy gracefully handles deleted clusters
2. **Error response generation**: Appropriate errors returned for missing clusters
3. **ODCDS behavior**: Discovery service properly removes deleted clusters
4. **Route resilience**: Virtual host continues to exist but routes fail appropriately