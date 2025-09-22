# Stateful Session Header Scenario

This scenario demonstrates the use of Envoy's stateful session header filter for maintaining session affinity. The header filter tracks session state by encoding upstream host addresses directly in request/response headers.

## Run the following scripts

You'll need to run these in separate tabs

Run the control-plane

```
./scripts/build_and_run_control_plane.sh
```

Run envoy

```
./scripts/run_envoy.sh
```

Run test backends

```
./scripts/start_test_services.sh
```

Update control-plane cache

```
go run cli/*.go -config scenarios/vhds-odcds-stateful-session-header/stateful-session-header-config.json -action update
```

Make client calls with persistent connection

```
go run pinger/*.go
```

## How it works

The stateful session header filter maintains session affinity by:
- **On request**: Checking for a `session-header` containing a base64-encoded upstream host address
- **On response**: Encoding the selected upstream host address into the `session-header` response header
- **Session persistence**: Subsequent requests with the same session header will be routed to the same upstream host

## Key differences from envelope filter

Unlike the envelope filter which embeds session context from the upstream server:
- Header filter directly encodes the upstream host address (e.g., `127.0.0.1:50053` becomes base64 encoded)
- Simpler approach focused purely on host affinity rather than application-level session state
- Does not require upstream server cooperation to generate session context

Press Ctrl+C to stop the pinger.