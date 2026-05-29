# On-Demand VHDS RecreateStream Crash

This scenario reproduces the production crash covered by
`VhdsOnDemandRecreateBufferedBodyTest.RecreateStreamWithBufferedBodyOverWatermark`.
The bug is present in the v1.37.2 Envoy build used for this investigation.

The important preconditions are:

1. HTTP/1 downstream traffic.
2. A 1048576 byte downstream buffer limit.
3. `envoy.filters.http.on_demand` before the router.
4. A static warmup host that loads the dynamic cluster without populating VHDS.
5. A POST body larger than the buffer limit sent to a missing VHDS host.
6. A VHDS update that resolves that host after the full request body has been buffered.

On unfixed Envoy, publishing `vhost.first` causes the on-demand VHDS callback to call
`recreateStream()`. The buffered body move drains the old watermark buffer after
`response_encoder_` has been nulled, and Envoy crashes in the low-watermark callback.

## Run

Start the control plane, Envoy, and test backends in separate terminals:

```bash
./scripts/build_and_run_control_plane.sh
ENVOY=/path/to/envoy ./scripts/run_envoy.sh
./scripts/start_test_services.sh
```

Then run the scenario:

```bash
./scenarios/vhds-ondemand-recreate-crash/reproduce.sh
```

Use the Envoy binary from the branch you are debugging. The stock Docker image in
`docker-compose.yml` may not contain the exact crash behavior.

## Manual Steps

```bash
go build -o bin/cli ./cli
./bin/cli --action update -config scenarios/vhds-ondemand-recreate-crash/initial-config.json

curl --fail --http1.1 \
  -H "Host: warmup.local" \
  -H "Content-Type: application/json" \
  -d '{"name":"Warmup"}' \
  http://localhost:10001/test/sayhello

perl -e 'print "{\"name\":\"", "a" x (1024 * 1024 + 4096), "\"}\n"' >/tmp/vhds-crash-body.json
curl --http1.1 -v \
  -H "Host: vhost.first" \
  -H "Content-Type: application/json" \
  --data-binary @/tmp/vhds-crash-body.json \
  http://localhost:10001/test/sayhello
```

While the last curl is waiting:

```bash
./bin/cli --action update -config scenarios/vhds-ondemand-recreate-crash/vhost-config.json
```
