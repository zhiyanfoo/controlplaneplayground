package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	http_conn "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	testv3 "github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	gRPCport        = 18000
	XDSHost         = "localhost"
	upstreamPort    = 50051 // Port of the test gRPC server
	listenerName    = "listener_0"
	listenerPort    = 10000
	routeName       = "local_route"
	clusterName     = "test_cluster"
	upstreamHost    = "127.0.0.1"
	nodeID          = "test-id"
	ListenerType    = resourcev3.ListenerType
	APITypePrefix   = "type.googleapis.com/envoy.config."
	fabricAuthority = ""
)

// --- Resource Generation Functions ---

func makeCluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       makeEndpoint(clusterName, upstreamHost, upstreamPort),
	}
}

func makeEndpoint(clusterName string, upstreamHost string, upstreamPort uint32) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol:      core.SocketAddress_TCP,
									Address:       upstreamHost,
									PortSpecifier: &core.SocketAddress_PortValue{PortValue: upstreamPort},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func makeRoute(routeName string, clusterName string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "local_service",
			Domains: []string{"*"}, // Match any domain
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"}, // Match any path
					Grpc:          &route.RouteMatch_GrpcRouteMatchOptions{},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{Cluster: clusterName},
					},
				},
			}},
		}},
	}
}

func makeHTTPListener(listenerName string, routeName string) *listener.Listener {
	hcm := &http_conn.HttpConnectionManager{
		CodecType:  http_conn.HttpConnectionManager_AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &http_conn.HttpConnectionManager_Rds{
			Rds: &http_conn.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: routeName,
			},
		},
		HttpFilters: []*http_conn.HttpFilter{{
			Name: wellknown.Router,
		}},
	}
	pbst, err := anypb.New(hcm)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol:      core.SocketAddress_TCP,
					Address:       "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{PortValue: listenerPort},
				},
			},
		},
		ApiListener: &listener.ApiListener{
			ApiListener: pbst, // Config for the HTTP connection manager
		},
	}
}

func makeConfigSource() *core.ConfigSource {
	return &core.ConfigSource{
		ResourceApiVersion: resourcev3.DefaultAPIVersion,
		ConfigSourceSpecifier: &core.ConfigSource_Ads{
			Ads: &core.AggregatedConfigSource{},
		},
	}
}

// MustAny converts a proto message to Any
func MustAny(p proto.Message) *anypb.Any {
	anypb, err := anypb.New(p)
	if err != nil {
		log.Fatalf("Failed to convert proto message to Any: %v", err)
	}
	return anypb
}

// --- Listener URI Helper ---

// URI defines the unique name of an xDS resource.
type URI string

func listenerUri(name, namespace string) string {
	listenerType := strings.TrimPrefix(ListenerType, APITypePrefix)
	uri := "xdstp://" + fabricAuthority + "/" + listenerType + "/" + name
	if namespace != "" {
		uri += "." + namespace
	}
	return uri
}

// --- Main Function ---

func main() {
	// Use LinearCache with listenerUri
	cache := xdscache.NewLinearCache(listenerUri("default-listener", "default-ns"))

	// Create resources
	l := makeHTTPListener(listenerName, routeName)
	r := makeRoute(routeName, clusterName)
	c := makeCluster(clusterName)
	e := makeEndpoint(clusterName, upstreamHost, upstreamPort)

	// Update cache - LinearCache UpdateResource expects (resourceType string, resource proto.Message)
	if err := cache.UpdateResource(resourcev3.ListenerType, l); err != nil {
		log.Fatalf("failed to update listener resource: %v", err)
	}
	if err := cache.UpdateResource(resourcev3.RouteType, r); err != nil {
		log.Fatalf("failed to update route resource: %v", err)
	}
	if err := cache.UpdateResource(resourcev3.ClusterType, c); err != nil {
		log.Fatalf("failed to update cluster resource: %v", err)
	}
	if err := cache.UpdateResource(resourcev3.EndpointType, e); err != nil {
		log.Fatalf("failed to update endpoint resource: %v", err)
	}

	log.Printf("Updated LinearCache with initial resources (LDS: %s, RDS: %s, CDS: %s, EDS: %s)", listenerName, routeName, clusterName, clusterName)

	ctx := context.Background()
	// Use testv3 callbacks for debugging
	cb := &testv3.Callbacks{Debug: true}
	srv := serverv3.NewServer(ctx, cache, cb)

	var grpcServer *grpc.Server
	grpcServer = grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", gRPCport))
	if err != nil {
		log.Fatal(err)
	}

	discoveryservice.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, srv)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, srv)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, srv)

	log.Printf("xDS control plane listening on %d\n", gRPCport)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait indefinitely
	<-ctx.Done()
	grpcServer.GracefulStop()
}
