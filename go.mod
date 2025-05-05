module controlplaneplayground

go 1.24.2

// We use the Datadog fork of the control-plane to include fixes not yet upstreamed
// As go-control-plane is the source of envoy protobuf files, we need to update it regularly
replace (
	github.com/envoyproxy/go-control-plane => github.com/DataDog/go-control-plane v0.13.4-0.20250422191546-6aa8b729178e
	github.com/envoyproxy/go-control-plane/envoy => github.com/DataDog/go-control-plane/envoy v1.32.4-0.20250422191546-6aa8b729178e
)

require (
	github.com/envoyproxy/go-control-plane v0.13.4
	github.com/envoyproxy/go-control-plane/envoy v1.32.4
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
)

require (
	cel.dev/expr v0.20.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250121191232-2f005788dc42 // indirect
	github.com/envoyproxy/go-control-plane/ratelimit v0.1.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
)
