# Stateful Session Envelope Scenario

This scenario demonstrates the use of Envoy's stateful session envelope filter for maintaining session affinity. The envelope filter tracks session state through a context initialized by the upstream server using a session header.

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
go run cli/*.go -config scenarios/vhds-odcds-stateful-session/stateful-session-config.json -action update
```

Make client calls with persistent connection

```
go run pinger/*.go
```

The stateful session envelope filter will maintain session affinity by:
- On response: Checking for session context and joining it with the upstream host
- On request: Checking for session context and stripping the upstream host
- Encoding the upstream host address and header value in base64 format

Press Ctrl+C to stop the pinger.