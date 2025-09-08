# Run the following scripts

You'll eed to run these in seperate tabs

Run the control-plane

```
./scripts/build_and_run_control_plane.sh
```

Run envoy

```
./run_envoy.sh
```

Run test backends

```
`./start_test_services.sh
```


Update control-plane cache

```
go run cli/*.go -config scenarios/vhds-odcds/vh1.json -action update
```

Make client call

```
./make_request.sh
```

The first client call will fail due to a bug in xds. The second one will second.
