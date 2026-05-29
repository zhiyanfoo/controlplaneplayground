#!/usr/bin/env bash
set -euo pipefail

SCENARIO_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
ROOT_DIR=$(cd -- "$SCENARIO_DIR/../.." && pwd)
CLI="$ROOT_DIR/bin/cli"

if [[ ! -x "$CLI" ]]; then
  (cd "$ROOT_DIR" && mkdir -p bin && go build -o bin/cli ./cli)
fi

VHDS_TYPE="type.googleapis.com/envoy.config.route.v3.VirtualHost"

echo "Clearing stale VHDS resources from previous runs..."
"$CLI" --action delete --type-url "$VHDS_TYPE" --name "vhds_crash_route/warmup.local" >/dev/null 2>&1 || true
"$CLI" --action delete --type-url "$VHDS_TYPE" --name "vhds_crash_route/vhost.first" >/dev/null 2>&1 || true

echo "Loading listener, RDS/VHDS, cluster, endpoint, and warmup vhost..."
"$CLI" --action update -config "$SCENARIO_DIR/initial-config.json"

echo "Waiting for Envoy listener localhost:10001..."
for _ in $(seq 1 50); do
  if (echo >/dev/tcp/127.0.0.1/10001) >/dev/null 2>&1; then
    break
  fi
  sleep 0.2
done

if ! (echo >/dev/tcp/127.0.0.1/10001) >/dev/null 2>&1; then
  echo "Envoy listener localhost:10001 did not become ready." >&2
  exit 1
fi

echo "Warming vhds_crash_cluster through ODCDS..."
curl --fail --http1.1 --silent --show-error \
  -H "Host: warmup.local" \
  -H "Content-Type: application/json" \
  -d '{"name":"Warmup"}' \
  "http://localhost:10001/test/sayhello" >/dev/null

body_file=$(mktemp /tmp/vhds-crash-body.XXXXXX.json)
trap 'rm -f "$body_file"' EXIT
perl -e 'print "{\"name\":\"", "a" x (1024 * 1024 + 4096), "\"}\n"' >"$body_file"

echo "Starting parked 1 MiB+ POST to missing VHDS host vhost.first..."
set +e
curl --http1.1 --verbose --max-time 30 \
  -H "Host: vhost.first" \
  -H "Content-Type: application/json" \
  -H "Expect:" \
  --data-binary @"$body_file" \
  "http://localhost:10001/test/sayhello" &
curl_pid=$!
set -e

sleep 1

echo "Publishing vhost.first over VHDS. Unfixed Envoy should crash here."
"$CLI" --action update -config "$SCENARIO_DIR/vhost-config.json"

set +e
wait "$curl_pid"
curl_status=$?
set -e

if [[ "$curl_status" -eq 0 ]]; then
  echo "Request completed. This is expected on a fixed Envoy."
else
  echo "curl exited with status $curl_status. Check Envoy logs for the recreateStream crash."
fi
