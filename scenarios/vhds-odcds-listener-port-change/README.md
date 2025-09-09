# VHDS-ODCDS Listener Port Change Test

This test demonstrates VHDS (Virtual Host Discovery Service) with ODCDS (On-Demand Cluster Discovery Service) behavior when a listener's port is changed.

## Test Scenario

This test verifies that when a listener's port is changed from 10000 to 10001, Envoy properly:
1. Updates the listener configuration via LDS (Listener Discovery Service)
2. Starts listening on the new port
3. Stops accepting connections on the old port
4. Maintains VHDS/ODCDS functionality on the new port

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

### 4. Load initial configuration (port 10000)
```
go run cli/*.go -config scenarios/vhds-odcds-listener-port-change/initial-config.json -action update
```

### 5. Start persistent client on port 10000 to observe connection behavior
```
go run pinger/*.go
```
*Note: First request may fail due to xDS timing - this is expected*

*Let this run in the background to observe how the persistent connection behaves during the port change*

### 6. Update configuration to port 10001 (while pinger is running)
```
go run cli/*.go -config scenarios/vhds-odcds-listener-port-change/updated-config.json -action update
```

*Observe in the pinger output how requests start failing when the port changes*

### 7. Start a new client on the updated port
```
go run pinger/*.go -addr localhost:10001
```
*Should succeed on port 10001 - demonstrates that VHDS/ODCDS works on the new port*

## Expected Behavior

- Initial requests succeed on port 10000
- After configuration update, port 10001 becomes active
- Port 10000 stops accepting new connections
- VHDS/ODCDS functionality works correctly on the new port
- Graceful transition without affecting existing backend connections

## Key Test Points

1. **Listener configuration updates**: Changes to listener port are applied correctly
2. **Port migration**: Traffic seamlessly moves to the new port
3. **Old port cleanup**: Previous port stops accepting connections
4. **VHDS/ODCDS preservation**: Virtual host and cluster discovery continue working on new port