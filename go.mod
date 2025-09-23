module controlplaneplayground

go 1.24.2

// We use the Datadog fork of the control-plane to include fixes not yet upstreamed
// As go-control-plane is the source of envoy protobuf files, we need to update it regularly
replace (
	github.com/envoyproxy/go-control-plane => github.com/DataDog/go-control-plane v0.0.4-zfs
	github.com/envoyproxy/go-control-plane/contrib => github.com/DataDog/go-control-plane/contrib v0.0.4-zfs
	github.com/envoyproxy/go-control-plane/envoy => github.com/DataDog/go-control-plane/envoy v0.0.4-zfs
)

require (
	github.com/envoyproxy/go-control-plane v0.13.4
	github.com/envoyproxy/go-control-plane/envoy v1.35.0
	github.com/google/uuid v1.6.0
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.7
)

require (
	cel.dev/expr v0.24.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/envoyproxy/go-control-plane/ratelimit v0.1.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250728155136-f173205681a0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250728155136-f173205681a0 // indirect
)
