# Run the following scripts

You'll need to run these in seperate tabs

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
`./scripts/start_test_services.sh
```


Update control-plane cache

```
go run cli/*.go -config scenarios/vhds-odcds-basic/basic-config.json -action update
```

Make client calls with persistent connection

```
go run pinger/*.go
```

The first client call will fail due to a bug in xds. The subsequent ones will succeed. Using the pinger allows you to observe that the persistent gRPC connection is maintained across XDS config changes. Press Ctrl+C to stop.
