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

### 5. Start persistent client to observe connection behavior
```
go run pinger/*.go
```
*Note: First request may fail due to xDS timing - this is expected*

*Let this run in the background to observe how the persistent connection behaves during the cluster change*

### 6. Update configuration to route to cluster B (while pinger is running)
```
go run cli/*.go -config scenarios/vhds-odcds-cluster-change/updated-config.json -action update
```

*Observe in the pinger output how the responses change from cluster A to cluster B without the gRPC connection being destroyed*

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