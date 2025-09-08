# VHDS-ODCDS Load Balancing Policy Change Test

This test demonstrates VHDS (Virtual Host Discovery Service) with ODCDS (On-Demand Cluster Discovery Service) behavior when a cluster's load balancing policy is updated.

## Test Scenario

This test verifies that when a cluster's load balancing policy is changed from ROUND_ROBIN to LEAST_REQUEST, Envoy properly:
1. Updates the cluster configuration on-demand via ODCDS
2. Applies the new load balancing policy
3. Handles the transition gracefully without service interruption

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

### 4. Load initial configuration (ROUND_ROBIN policy)
```
go run cli/*.go -config scenarios/vhds-odcds-lb-policy-change/initial-config.json -action update
```

### 5. Make initial requests
```
./scripts/make_request.sh
```
*Note: First request may fail due to xDS timing - this is expected*

```
./scripts/make_request.sh
```
*Should succeed with ROUND_ROBIN load balancing*

### 6. Update configuration to LEAST_REQUEST policy
```
go run cli/*.go -config scenarios/vhds-odcds-lb-policy-change/updated-config.json -action update
```

### 7. Make requests after policy change
```
./scripts/make_request.sh
```
*Should succeed with LEAST_REQUEST load balancing*

## Expected Behavior

- Initial requests use ROUND_ROBIN load balancing policy
- After configuration update, new requests use LEAST_REQUEST load balancing policy
- ODCDS automatically updates the cluster configuration when changed
- No service interruption during the load balancing policy transition

## Key Test Points

1. **Cluster configuration updates**: Changes to cluster definition are applied correctly
2. **Load balancing policy changes**: Policy updates take effect immediately
3. **ODCDS behavior**: Cluster updates are discovered and applied on-demand
4. **Graceful transition**: No service interruption during policy changes