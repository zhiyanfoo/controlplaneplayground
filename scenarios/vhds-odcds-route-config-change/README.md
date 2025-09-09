# VHDS-ODCDS Route Configuration Change Test (Per-Route ODCDS)

This test demonstrates VHDS (Virtual Host Discovery Service) with ODCDS (On-Demand Cluster Discovery Service) behavior when ODCDS is configured at the route level using `typedPerFilterConfig`, and the route configuration is subsequently changed.

## Test Scenario

This test verifies the behavior when ODCDS configuration is set as `typedPerFilterConfig` on individual routes rather than at the HTTP filter level. When the route configuration is updated (while keeping the ODCDS config), this scenario is known to trigger a bug in Envoy that leads to a crash.

**⚠️ WARNING: This test is expected to cause Envoy to crash due to a known bug when ODCDS is configured per-route and the route is updated.**

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

### 4. Load initial configuration (ODCDS configured per-route)
```
go run cli/*.go -config scenarios/vhds-odcds-route-config-change/initial-config.json -action update
```

### 5. Start persistent client to observe connection behavior
```
go run pinger/*.go
```
*Note: First request may fail due to xDS timing - this is expected*

*Let this run in the background to observe the crash behavior*

### 6. Update route configuration (while pinger is running) 
```
go run cli/*.go -config scenarios/vhds-odcds-route-config-change/updated-config.json -action update
```

*This should trigger the Envoy crash due to the bug with per-route ODCDS configuration updates*

## Expected Behavior

- Initial requests work normally with ODCDS configured at the route level
- After configuration update, **Envoy crashes** due to the bug with per-route ODCDS handling
- The pinger client will show connection failures after the crash

## Key Test Points

1. **Per-route ODCDS configuration**: ODCDS is configured using `typedPerFilterConfig` on the route instead of at the HTTP filter level
2. **Route update with per-route ODCDS**: Demonstrates the bug when route configuration changes while ODCDS is configured per-route
3. **Crash reproduction**: Successfully reproduces the known Envoy bug for debugging purposes
4. **Configuration difference**: Shows the behavior difference between filter-level and route-level ODCDS configuration

## Configuration Details

- **Initial config**: Route matches "/" with timeout default, ODCDS configured via `typedPerFilterConfig`
- **Updated config**: Route matches "/" with timeout "10s", same ODCDS `typedPerFilterConfig` 
- **Key difference**: ODCDS configuration is at route level, not HTTP filter level