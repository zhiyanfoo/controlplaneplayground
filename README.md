# To run

You'll need to run multiple components in seperate shells

The control plane that serves the xds configuration
```
./scripts/build_and_run_control_plane.sh
```

Envoy instance
```
scripts/run_envoy.sh
```

Grpc test server
```
./scripts/build_and_run_test_server.sh
```

The make a grpc request with

```
./scripts/make_request.sh
```

The request will be made to Envoy, which should forward it to test-server.
Envoy should get it's configuration from the control-plane.


