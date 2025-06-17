package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	file_accesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	upstreamhttp "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	debugLogFilename = "xds_debug.log"
	gRPCport         = 18000
	XDSHost          = "localhost"
	upstreamPort     = 50051 // Port of the test gRPC server
	listenerName     = "listener_0"
	listenerPort     = 10000
	routeName        = "local_route"
	clusterName      = "test_cluster"
	upstreamHost     = "127.0.0.1"
	nodeID           = "test-id" // Node ID served by this control plane

	// Resource Type URLs
	ListenerType    = resourcev3.ListenerType
	RouteType       = resourcev3.RouteType
	ClusterType     = resourcev3.ClusterType
	EndpointType    = resourcev3.EndpointType
	VirtualHostType = resourcev3.VirtualHostType
	APITypePrefix   = "type.googleapis.com/envoy.config."
	fabricAuthority = ""
)

// --- Resource Generation Functions ---

func makeCluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: makeConfigSource(),
		},
		LbPolicy: cluster.Cluster_ROUND_ROBIN,
		TypedExtensionProtocolOptions: map[string]*anypb.Any{
			"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": MustAny(&upstreamhttp.HttpProtocolOptions{
				UpstreamProtocolOptions: &upstreamhttp.HttpProtocolOptions_ExplicitHttpConfig_{
					ExplicitHttpConfig: &upstreamhttp.HttpProtocolOptions_ExplicitHttpConfig{
						ProtocolConfig: &upstreamhttp.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
							Http2ProtocolOptions: &core.Http2ProtocolOptions{},
						},
					},
				},
			}),
		},
	}
}

func makeVirtualHost(virtualHostName string, domains []string, clusterName string) *route.VirtualHost {
	return &route.VirtualHost{
		Name:    virtualHostName,
		Domains: domains,
		Routes: []*route.Route{{
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{Cluster: clusterName},
				},
			},
		}},
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

func makeRoute(routeName string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: routeName,
		Vhds: &route.Vhds{
			ConfigSource: makeVhdsConfigSource(),
		},
	}
}

func makeHTTPListener(listenerName string, routeConfigName string) *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(),
				RouteConfigName: routeConfigName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       "http-router",
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
		AccessLog: []*accesslog.AccessLog{
			{
				Name: "envoy.access_loggers.file",
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: MustAny(&file_accesslog.FileAccessLog{
						Path: "/dev/stdout",
						AccessLogFormat: &file_accesslog.FileAccessLog_LogFormat{
							LogFormat: &core.SubstitutionFormatString{
								Format: &core.SubstitutionFormatString_TextFormatSource{
									TextFormatSource: &core.DataSource{
										Specifier: &core.DataSource_InlineString{
											InlineString: "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% %GRPC_STATUS%(%GRPC_STATUS_NUMBER%) %BYTES_SENT% %DURATION% CTD %CONNECTION_TERMINATION_DETAILS% URAC %UPSTREAM_REQUEST_ATTEMPT_COUNT% DWBS %DOWNSTREAM_WIRE_BYTES_SENT% USWBR %UPSTREAM_WIRE_BYTES_RECEIVED% UTFR %UPSTREAM_TRANSPORT_FAILURE_REASON% UH %UPSTREAM_HOST% UC %UPSTREAM_CLUSTER% GRPC_MSG %RESP(grpc-message)%\n",
										},
									},
								},
							},
						},
					}),
				},
			},
		},
	}
	pbst, err := anypb.New(manager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: listenerPort,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: "http-connection-manager",
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}

// Reverted makeConfigSource to its original simpler form
func makeConfigSource() *core.ConfigSource {
	return &core.ConfigSource{
		ResourceApiVersion: resourcev3.DefaultAPIVersion,
		ConfigSourceSpecifier: &core.ConfigSource_Ads{
			Ads: &core.AggregatedConfigSource{},
		},
	}
}

// New function for VHDS ConfigSource using Delta GRPC (non-ADS)
func makeVhdsConfigSource() *core.ConfigSource {
	return &core.ConfigSource{
		ResourceApiVersion: resourcev3.DefaultAPIVersion,
		ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
			ApiConfigSource: &core.ApiConfigSource{
				ApiType:             core.ApiConfigSource_DELTA_GRPC,
				TransportApiVersion: resourcev3.DefaultAPIVersion,
				GrpcServices: []*core.GrpcService{{
					TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"}, // Using the defined clusterName constant
					},
				}},
				// SetInitialFetchTimeout: durationpb.New(1 * time.Second), // Optional: can be added if needed
			},
		},
		InitialFetchTimeout: durationpb.New(time.Second),
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

// --- Custom Callbacks for Logging Requests and Responses ---

// requestResponseLogger implements serverv3.Callbacks for detailed xDS message logging.
type requestResponseLogger struct {
	debugFileLogger *log.Logger // Logger for writing detailed output to a file
}

// OnStreamOpen is called once an xDS stream is open.
func (l *requestResponseLogger) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	log.Printf("CONSOLE: OnStreamOpen ID[%d] Type[%s]", id, typ)
	// No detailed protobuf to write to file for this event
	return nil
}

// OnStreamClosed is called immediately prior to closing an xDS stream.
func (l *requestResponseLogger) OnStreamClosed(id int64, node *core.Node) {
	var nodeID string
	if node != nil {
		nodeID = node.GetId()
	}
	log.Printf("CONSOLE: OnStreamClosed ID[%d] NodeID[%s]", id, nodeID)
	// No detailed protobuf to write to file for this event
}

// OnStreamRequest logs incoming SotW DiscoveryRequests.
func (l *requestResponseLogger) OnStreamRequest(streamID int64, req *discoveryservice.DiscoveryRequest) error {
	log.Printf("CONSOLE: OnStreamRequest ID[%d] Type[%s] Node[%s] Version[%s] Nonce[%s] Resources[%d]",
		streamID, req.GetTypeUrl(), req.GetNode().GetId(), req.GetVersionInfo(), req.GetResponseNonce(), len(req.GetResourceNames()))
	if l.debugFileLogger != nil {
		l.debugFileLogger.Printf("FILE DEBUG: OnStreamRequest ID[%d]:\n%s\n---END REQUEST---", streamID, req.String())
	}
	return nil
}

// OnStreamResponse logs outgoing SotW DiscoveryResponses.
func (l *requestResponseLogger) OnStreamResponse(ctx context.Context, streamID int64, req *discoveryservice.DiscoveryRequest, resp *discoveryservice.DiscoveryResponse) {
	log.Printf("CONSOLE: OnStreamResponse ID[%d] Type[%s] Version[%s] Nonce[%s] Resources[%d] for Request Node[%s]",
		streamID, resp.GetTypeUrl(), resp.GetVersionInfo(), resp.GetNonce(), len(resp.GetResources()), req.GetNode().GetId())
	if l.debugFileLogger != nil {
		l.debugFileLogger.Printf("FILE DEBUG: OnStreamResponse ID[%d] for Request Node[%s] Type[%s]:\n%s\n---END RESPONSE---", streamID, req.GetNode().GetId(), req.GetTypeUrl(), resp.String())
	}
}

// OnDeltaStreamOpen is called once an xDS Delta stream is open.
func (l *requestResponseLogger) OnDeltaStreamOpen(ctx context.Context, id int64, typ string) error {
	log.Printf("CONSOLE: OnDeltaStreamOpen ID[%d] Type[%s]", id, typ)
	// No detailed protobuf to write to file for this event
	return nil
}

// OnDeltaStreamClosed is called immediately prior to closing an xDS Delta stream.
func (l *requestResponseLogger) OnDeltaStreamClosed(id int64, node *core.Node) {
	var nodeID string
	if node != nil {
		nodeID = node.GetId()
	}
	log.Printf("CONSOLE: OnDeltaStreamClosed ID[%d] NodeID[%s]", id, nodeID)
	// No detailed protobuf to write to file for this event
}

// OnStreamDeltaRequest logs incoming DeltaDiscoveryRequests.
// Signature matches older go-control-plane interface expected by the linter.
func (l *requestResponseLogger) OnStreamDeltaRequest(streamID int64, req *discoveryservice.DeltaDiscoveryRequest) error {
	log.Printf("CONSOLE: OnStreamDeltaRequest ID[%d] Type[%s] Node[%s] RespNonce[%s] Sub[%d] Unsub[%d]",
		streamID, req.GetTypeUrl(), req.GetNode().GetId(), req.GetResponseNonce(), len(req.GetResourceNamesSubscribe()), len(req.GetResourceNamesUnsubscribe()))
	if l.debugFileLogger != nil {
		l.debugFileLogger.Printf("FILE DEBUG: OnStreamDeltaRequest ID[%d] TypeURL[%s]:\n%s\n---END REQUEST---", streamID, req.GetTypeUrl(), req.String())
	}
	return nil
}

// OnStreamDeltaResponse logs outgoing DeltaDiscoveryResponses.
// Name matches older go-control-plane interface.
func (l *requestResponseLogger) OnStreamDeltaResponse(streamID int64, req *discoveryservice.DeltaDiscoveryRequest, resp *discoveryservice.DeltaDiscoveryResponse) {
	log.Printf("CONSOLE: OnStreamDeltaResponse ID[%d] Type[%s] SysVer[%s] Nonce[%s] ResourcesSent[%d] ResourcesRemoved[%d] for ReqNode[%s]",
		streamID, resp.GetTypeUrl(), resp.GetSystemVersionInfo(), resp.GetNonce(), len(resp.GetResources()), len(resp.GetRemovedResources()), req.GetNode().GetId())
	if l.debugFileLogger != nil {
		l.debugFileLogger.Printf("FILE DEBUG: OnStreamDeltaResponse ID[%d] for Request Node[%s] TypeURL[%s]:\n%s\n---END RESPONSE---", streamID, req.GetNode().GetId(), req.GetTypeUrl(), resp.String())
	}
}

// OnFetchRequest logs incoming Fetch DiscoveryRequests.
func (l *requestResponseLogger) OnFetchRequest(ctx context.Context, req *discoveryservice.DiscoveryRequest) error {
	log.Printf("CONSOLE: OnFetchRequest Node[%s] Type[%s] Version[%s] Nonce[%s] Resources[%d]",
		req.GetNode().GetId(), req.GetTypeUrl(), req.GetVersionInfo(), req.GetResponseNonce(), len(req.GetResourceNames()))
	if l.debugFileLogger != nil {
		l.debugFileLogger.Printf("FILE DEBUG: OnFetchRequest Node[%s] TypeURL[%s]:\n%s\n---END REQUEST---", req.GetNode().GetId(), req.GetTypeUrl(), req.String())
	}
	return nil
}

// OnFetchResponse logs outgoing Fetch DiscoveryResponses.
func (l *requestResponseLogger) OnFetchResponse(req *discoveryservice.DiscoveryRequest, resp *discoveryservice.DiscoveryResponse) {
	log.Printf("CONSOLE: OnFetchResponse Type[%s] Version[%s] Nonce[%s] Resources[%d] for Request Node[%s]",
		resp.GetTypeUrl(), resp.GetVersionInfo(), resp.GetNonce(), len(resp.GetResources()), req.GetNode().GetId())
	if l.debugFileLogger != nil {
		l.debugFileLogger.Printf("FILE DEBUG: OnFetchResponse for Request Node[%s] TypeURL[%s]:\n%s\n---END RESPONSE---", req.GetNode().GetId(), req.GetTypeUrl(), resp.String())
	}
}

// --- Main Function ---

func main() {
	// Setup file logger for detailed debug messages
	debugFile, err := os.OpenFile(debugLogFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open debug log file %s: %v", debugLogFilename, err)
	}
	defer debugFile.Close()
	debugFileLogger := log.New(debugFile, "", log.LstdFlags) // Using standard log flags, no prefix from logger itself

	log.Printf("Control plane starting. Detailed xDS debug logs will be written to %s", debugLogFilename)

	// --- Initialize Caches ---
	listenerCache := xdscache.NewLinearCache(ListenerType)
	clusterCache := xdscache.NewLinearCache(ClusterType)
	routeCache := xdscache.NewLinearCache(RouteType)
	endpointCache := xdscache.NewLinearCache(EndpointType)
	virtualHostCache := xdscache.NewLinearCache(VirtualHostType)

	// --- Create Resources ---
	// routeName is the name of the RouteConfiguration object itself
	// For VHDS, the VirtualHost resource name must be <RouteConfiguration_Name>/<Authority_Header_Value>
	const expectedAuthority = "localhost:10000" // This is what grpcurl will use to call Envoy
	virtualHostName := routeName + "/" + expectedAuthority

	l := makeHTTPListener(listenerName, routeName) // Listener refers to RouteConfiguration named routeName ("local_route")
	rc := makeRoute(routeName)                     // RouteConfiguration delegates to VHDS
	// The VirtualHost must match what Envoy will request: <RouteConfigName>/<HostHeader>
	vh := makeVirtualHost(virtualHostName, []string{expectedAuthority, "*"}, clusterName) // VirtualHost resource

	// Create original cluster
	c := makeCluster(clusterName)
	e := makeEndpoint(clusterName, upstreamHost, upstreamPort)

	// Create three additional clusters
	c1 := makeCluster("test_cluster_1")
	c2 := makeCluster("test_cluster_2")
	c3 := makeCluster("test_cluster_3")

	// Create corresponding endpoints for each additional cluster
	e1 := makeEndpoint("test_cluster_1", upstreamHost, upstreamPort)
	e2 := makeEndpoint("test_cluster_2", upstreamHost, upstreamPort)
	e3 := makeEndpoint("test_cluster_3", upstreamHost, upstreamPort)

	// --- Populate Linear Caches ---
	if err := listenerCache.UpdateResource(listenerName, l); err != nil {
		log.Fatalf("failed to update listener resource in listener cache: %v", err)
	}

	if err := routeCache.UpdateResource(routeName, rc); err != nil {
		log.Fatalf("failed to update route configuration resource in route cache: %v", err)
	}

	if err := virtualHostCache.UpdateResource(virtualHostName, vh); err != nil { // Store VirtualHost for VHDS
		log.Fatalf("failed to update virtual host resource in VHDS cache: %v", err)
	}

	// Update cluster cache with original cluster
	if err := clusterCache.UpdateResource(clusterName, c); err != nil {
		log.Fatalf("failed to update cluster resource in cluster cache: %v", err)
	}

	// Update cluster cache with additional clusters
	if err := clusterCache.UpdateResource("test_cluster_1", c1); err != nil {
		log.Fatalf("failed to update cluster resource test_cluster_1 in cluster cache: %v", err)
	}
	if err := clusterCache.UpdateResource("test_cluster_2", c2); err != nil {
		log.Fatalf("failed to update cluster resource test_cluster_2 in cluster cache: %v", err)
	}
	if err := clusterCache.UpdateResource("test_cluster_3", c3); err != nil {
		log.Fatalf("failed to update cluster resource test_cluster_3 in cluster cache: %v", err)
	}

	// Update endpoint cache with original endpoint
	if err := endpointCache.UpdateResource(clusterName, e); err != nil {
		log.Fatalf("failed to update endpoint resource in endpoint cache: %v", err)
	}

	// Update endpoint cache with additional endpoints
	if err := endpointCache.UpdateResource("test_cluster_1", e1); err != nil {
		log.Fatalf("failed to update endpoint resource test_cluster_1 in endpoint cache: %v", err)
	}
	if err := endpointCache.UpdateResource("test_cluster_2", e2); err != nil {
		log.Fatalf("failed to update endpoint resource test_cluster_2 in endpoint cache: %v", err)
	}
	if err := endpointCache.UpdateResource("test_cluster_3", e3); err != nil {
		log.Fatalf("failed to update endpoint resource test_cluster_3 in endpoint cache: %v", err)
	}

	log.Printf("Updated Linear caches (LDS: %s, RDS: %s, VHDS: %s, CDS: %s,test_cluster_1,2,3, EDS: %s,test_cluster_1,2,3)", listenerName, routeName, virtualHostName, clusterName, clusterName)

	// --- Create MuxCache for ADS (LDS, RDS, CDS, EDS) ---
	muxCache := &xdscache.MuxCache{
		Classify: func(req *xdscache.Request) string {
			return req.TypeUrl
		},
		ClassifyDelta: func(req *xdscache.DeltaRequest) string {
			return req.TypeUrl
		},
		Caches: map[string]xdscache.Cache{
			ListenerType:    listenerCache,
			RouteType:       routeCache,
			ClusterType:     clusterCache,
			EndpointType:    endpointCache,
			VirtualHostType: virtualHostCache,
		},
	}

	ctx := context.Background()
	// Use custom logger implementing serverv3.Callbacks
	cb := &requestResponseLogger{debugFileLogger: debugFileLogger}

	// Server for ADS and VHDS
	srv := serverv3.NewServer(ctx, muxCache, cb)

	// --- Start gRPC Server ---
	var grpcServer *grpc.Server
	// Remove or keep gRPC interceptor based on needs. For now, removing it as we have xDS layer logging.
	grpcServer = grpc.NewServer()
	// Listen on loopback only
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", gRPCport))
	if err != nil {
		log.Fatal(err)
	}

	// Register ADS server
	discoveryservice.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	// Register dedicated VHDS server
	routeservice.RegisterVirtualHostDiscoveryServiceServer(grpcServer, srv)

	log.Printf("xDS control plane: ADS and VHDS (Delta GRPC) services listening on %d\n", gRPCport)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait indefinitely
	<-ctx.Done()
	grpcServer.GracefulStop()
}
