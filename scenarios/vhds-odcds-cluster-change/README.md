# VHDS-ODCDS Cluster Change Test

This test demonstrates VHDS (Virtual Host Discovery Service) with ODCDS (On-Demand Cluster Discovery Service) behavior when the cluster name referenced in a route is changed to point to a different cluster.

## Test Scenario

This test verifies that when a route's cluster reference is updated from one cluster to another, Envoy properly:
1. Discovers and loads the new cluster on-demand via ODCDS
2. Routes traffic to the new cluster 
3. Handles the transition gracefully

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

### 4. Load initial configuration (routes to cluster A)
```
go run cli/*.go -config scenarios/vhds-odcds-cluster-change/initial-config.json -action update
```

### 5. Make initial request
```
./scripts/make_request.sh
```
*Note: First request may fail due to xDS timing - this is expected*

### 6. Make second request (should succeed with cluster A response)
```
./scripts/make_request.sh
```

### 7. Update configuration to route to cluster B
```
go run cli/*.go -config scenarios/vhds-odcds-cluster-change/updated-config.json -action update
```

### 8. Make request after cluster change
```
./scripts/make_request.sh
```
*Should now route to cluster B and return different response*

## Expected Behavior

- Initial requests route to the original cluster (basic_grpc_cluster)
- After configuration update, new requests route to the updated cluster (alternative_grpc_cluster)  
- ODCDS automatically discovers and loads the new cluster when needed
- Traffic seamlessly transitions to the new backend without service interruption

## Key Test Points

1. **On-demand cluster discovery**: New cluster is discovered only when referenced by a route
2. **Route update handling**: Virtual host route updates are applied correctly
3. **Traffic switching**: Requests are routed to the correct cluster after updates
4. **Graceful transition**: No service interruption during cluster changes