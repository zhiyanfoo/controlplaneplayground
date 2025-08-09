package main

import (
	"context"
	"controlplaneplayground/pb"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	file_accesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	ondemand "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/on_demand/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	upstreamhttp "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	debugLogFilename = "xds_debug.log"
	gRPCport         = 18000
	HTTPport         = 8080 // HTTP server for cache display
	XDSHost          = "localhost"
	upstreamPort     = 50051 // Port of the test gRPC server
	listenerName     = "listener_0"
	listenerPort     = 10000
	routeName        = "local_route"
	clusterName      = "xdstp:///envoy.config.cluster.v3.Cluster/test_cluster"
	upstreamHost     = "127.0.0.1"
	nodeID           = "test-id" // Node ID served by this control plane

	// HTTP/1.1 specific constants
	http1ListenerName = "http1_listener_0"
	http1ListenerPort = 10001
	http1RouteName    = "http1_route"
	http1ClusterName  = "xdstp:///envoy.config.cluster.v3.Cluster/http1_test_cluster"
	http1UpstreamPort = 50052 // HTTP server port

	// Dynamic vhost specific constants
	dynamicRouteName    = "dynamic_route"
	dynamicListenerPort = 10002
	dynamicAuthority    = "localhost:10002"

	// Access log format constants
	HTTP1AccessLogFormat = "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% %BYTES_SENT% %DURATION% %UPSTREAM_HOST% %UPSTREAM_CLUSTER%\n"
	HTTP2AccessLogFormat = "[%START_TIME%] \"%REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\" %RESPONSE_CODE% %RESPONSE_FLAGS% %GRPC_STATUS%(%GRPC_STATUS_NUMBER%) %BYTES_SENT% %DURATION% CTD %CONNECTION_TERMINATION_DETAILS% URAC %UPSTREAM_REQUEST_ATTEMPT_COUNT% DWBS %DOWNSTREAM_WIRE_BYTES_SENT% USWBR %UPSTREAM_WIRE_BYTES_RECEIVED% UTFR %UPSTREAM_TRANSPORT_FAILURE_REASON% UH %UPSTREAM_HOST% UC %UPSTREAM_CLUSTER% GRPC_MSG %RESP(grpc-message)%\n"

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

// ListenerConfig holds configuration for HTTP listeners
type ListenerConfig struct {
	Name            string
	RouteConfigName string
	StatPrefix      string
	Port            uint32
	AccessLogFormat string
}

func makeCluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: makeADSConfigSource(),
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

func makeHttp1Cluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: makeADSConfigSource(),
		},
		LbPolicy: cluster.Cluster_ROUND_ROBIN,
		TypedExtensionProtocolOptions: map[string]*anypb.Any{
			"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": MustAny(&upstreamhttp.HttpProtocolOptions{
				UpstreamProtocolOptions: &upstreamhttp.HttpProtocolOptions_ExplicitHttpConfig_{
					ExplicitHttpConfig: &upstreamhttp.HttpProtocolOptions_ExplicitHttpConfig{
						ProtocolConfig: &upstreamhttp.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{
							HttpProtocolOptions: &core.Http1ProtocolOptions{},
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
		TypedPerFilterConfig: map[string]*anypb.Any{
			"envoy.filters.http.on_demand": MustAny(&ondemand.PerRouteConfig{
				Odcds: &ondemand.OnDemandCds{
					Source:  makeADSConfigSource(),
					Timeout: durationpb.New(1 * time.Second),
				},
			}),
		},
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

func makeOdcdsConfig() (*anypb.Any, error) {
	return anypb.New(&ondemand.OnDemand{
		Odcds:  &ondemand.OnDemandCds{
			Source: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
				ResourceApiVersion: core.ApiVersion_V3,
			},
			Timeout: durationpb.New(time.Second),
		},
	})
}

func makeHTTPListener(config ListenerConfig) *listener.Listener {
	routerConfig, err := anypb.New(&router.Router{})
	if err != nil {
		panic(err)
	}


	odcdsConfig, err :=  makeOdcdsConfig()
	if err != nil {
		panic(err)
	}

	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: config.StatPrefix,
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeADSConfigSource(),
				RouteConfigName: config.RouteConfigName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{
			{
				Name:       "envoy.filters.http.on_demand",
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: odcdsConfig},
			},
			{
				Name:       "http-router",
				ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
			},
		},
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
											InlineString: config.AccessLogFormat,
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
		Name: config.Name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: config.Port,
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

// Reverted makeADSConfigSource to its original simpler form
func makeADSConfigSource() *core.ConfigSource {
	return &core.ConfigSource{
		ResourceApiVersion: resourcev3.DefaultAPIVersion,
		ConfigSourceSpecifier: &core.ConfigSource_Ads{
			Ads: &core.AggregatedConfigSource{},
		},
		InitialFetchTimeout: durationpb.New(1 * time.Second),
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

// CacheDisplayHandler handles HTTP requests to display cache contents
type CacheDisplayHandler struct {
	listenerCache    *xdscache.LinearCache
	clusterCache     *xdscache.LinearCache
	routeCache       *xdscache.LinearCache
	endpointCache    *xdscache.LinearCache
	virtualHostCache *xdscache.LinearCache
}

// NewCacheDisplayHandler creates a new cache display handler
func NewCacheDisplayHandler(listenerCache, clusterCache, routeCache, endpointCache, virtualHostCache *xdscache.LinearCache) *CacheDisplayHandler {
	return &CacheDisplayHandler{
		listenerCache:    listenerCache,
		clusterCache:     clusterCache,
		routeCache:       routeCache,
		endpointCache:    endpointCache,
		virtualHostCache: virtualHostCache,
	}
}

// ResourceInfo holds information about a cached resource
type ResourceInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	JSONData string `json:"json_data"`
}

// CacheState holds the complete state of all caches
type CacheState struct {
	Listeners    []ResourceInfo `json:"listeners"`
	Clusters     []ResourceInfo `json:"clusters"`
	Routes       []ResourceInfo `json:"routes"`
	Endpoints    []ResourceInfo `json:"endpoints"`
	VirtualHosts []ResourceInfo `json:"virtual_hosts"`
}

// ServeHTTP handles the cache display request
func (h *CacheDisplayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cacheState := h.getCacheState()

	// Check if JSON output is requested
	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cacheState)
		return
	}

	// Serve HTML page
	w.Header().Set("Content-Type", "text/html")
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>XDS Cache Contents</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .resource-type { margin: 20px 0; }
        .resource-type h2 { color: #333; border-bottom: 2px solid #ddd; padding-bottom: 5px; }
        .resource { margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 5px; }
        .resource-name { font-weight: bold; color: #0066cc; }
        .resource-json { 
            background: #f5f5f5; 
            padding: 10px; 
            margin: 5px 0; 
            border-radius: 3px; 
            overflow-x: auto;
            white-space: pre-wrap;
            font-family: monospace;
            font-size: 12px;
        }
        .refresh-btn { 
            background: #4CAF50; 
            color: white; 
            padding: 10px 20px; 
            text-decoration: none; 
            border-radius: 5px; 
            display: inline-block; 
            margin-bottom: 20px;
        }
        .json-btn { 
            background: #2196F3; 
            color: white; 
            padding: 5px 10px; 
            text-decoration: none; 
            border-radius: 3px; 
            display: inline-block; 
            margin-left: 10px;
        }
    </style>
</head>
<body>
    <h1>XDS Cache Contents</h1>
    <a href="/" class="refresh-btn">Refresh</a>
    <a href="/?format=json" class="json-btn">View as JSON</a>

    <div class="resource-type">
        <h2>Listeners ({{len .Listeners}})</h2>
        {{range .Listeners}}
        <div class="resource">
            <div class="resource-name">{{.Name}}</div>
            <div class="resource-json">{{.JSONData}}</div>
        </div>
        {{end}}
    </div>

    <div class="resource-type">
        <h2>Clusters ({{len .Clusters}})</h2>
        {{range .Clusters}}
        <div class="resource">
            <div class="resource-name">{{.Name}}</div>
            <div class="resource-json">{{.JSONData}}</div>
        </div>
        {{end}}
    </div>

    <div class="resource-type">
        <h2>Routes ({{len .Routes}})</h2>
        {{range .Routes}}
        <div class="resource">
            <div class="resource-name">{{.Name}}</div>
            <div class="resource-json">{{.JSONData}}</div>
        </div>
        {{end}}
    </div>

    <div class="resource-type">
        <h2>Endpoints ({{len .Endpoints}})</h2>
        {{range .Endpoints}}
        <div class="resource">
            <div class="resource-name">{{.Name}}</div>
            <div class="resource-json">{{.JSONData}}</div>
        </div>
        {{end}}
    </div>

    <div class="resource-type">
        <h2>Virtual Hosts ({{len .VirtualHosts}})</h2>
        {{range .VirtualHosts}}
        <div class="resource">
            <div class="resource-name">{{.Name}}</div>
            <div class="resource-json">{{.JSONData}}</div>
        </div>
        {{end}}
    </div>
</body>
</html>
`
	t := template.Must(template.New("cache").Parse(tmpl))
	t.Execute(w, cacheState)
}

// getCacheState retrieves the current state of all caches
func (h *CacheDisplayHandler) getCacheState() CacheState {
	state := CacheState{
		Listeners:    []ResourceInfo{},
		Clusters:     []ResourceInfo{},
		Routes:       []ResourceInfo{},
		Endpoints:    []ResourceInfo{},
		VirtualHosts: []ResourceInfo{},
	}

	// Get resources from all caches
	listenerResources := h.listenerCache.GetResources()
	clusterResources := h.clusterCache.GetResources()
	routeResources := h.routeCache.GetResources()
	endpointResources := h.endpointCache.GetResources()
	vhostResources := h.virtualHostCache.GetResources()

	// Process listeners
	for name, resource := range listenerResources {
		if listener, ok := resource.(*listener.Listener); ok {
			jsonData, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(listener)
			state.Listeners = append(state.Listeners, ResourceInfo{
				Name:     name,
				Type:     "Listener",
				JSONData: string(jsonData),
			})
		}
	}

	// Process clusters
	for name, resource := range clusterResources {
		if cluster, ok := resource.(*cluster.Cluster); ok {
			jsonData, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(cluster)
			state.Clusters = append(state.Clusters, ResourceInfo{
				Name:     name,
				Type:     "Cluster",
				JSONData: string(jsonData),
			})
		}
	}

	// Process routes
	for name, resource := range routeResources {
		if route, ok := resource.(*route.RouteConfiguration); ok {
			jsonData, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(route)
			state.Routes = append(state.Routes, ResourceInfo{
				Name:     name,
				Type:     "Route",
				JSONData: string(jsonData),
			})
		}
	}

	// Process endpoints
	for name, resource := range endpointResources {
		if endpoint, ok := resource.(*endpoint.ClusterLoadAssignment); ok {
			jsonData, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(endpoint)
			state.Endpoints = append(state.Endpoints, ResourceInfo{
				Name:     name,
				Type:     "Endpoint",
				JSONData: string(jsonData),
			})
		}
	}

	// Process virtual hosts
	for name, resource := range vhostResources {
		if vhost, ok := resource.(*route.VirtualHost); ok {
			jsonData, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(vhost)
			state.VirtualHosts = append(state.VirtualHosts, ResourceInfo{
				Name:     name,
				Type:     "VirtualHost",
				JSONData: string(jsonData),
			})
		}
	}

	return state
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

// xdsCallbackManager implements serverv3.Callbacks and manages xDS callbacks with access to caches
type xdsCallbackManager struct {
	*requestResponseLogger
	virtualHostCache *xdscache.LinearCache
	clusterCache     *xdscache.LinearCache
	nodeClusterCache map[string]map[string]struct{} // Maps nodeID to set of clusters
	mu               sync.RWMutex                   // Protects nodeClusterCache
}

// ClusterList holds the list of clusters for a node
type ClusterList struct {
	Clusters []string
}

// NewXdsCallbackManager creates a new xdsCallbackManager instance
func NewXdsCallbackManager(logger *requestResponseLogger, vhCache, clCache *xdscache.LinearCache) *xdsCallbackManager {
	return &xdsCallbackManager{
		requestResponseLogger: logger,
		virtualHostCache:      vhCache,
		clusterCache:          clCache,
		nodeClusterCache:      make(map[string]map[string]struct{}),
	}
}

// OnStreamOpen is called once an xDS stream is open.
func (m *xdsCallbackManager) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	return m.requestResponseLogger.OnStreamOpen(ctx, id, typ)
}

// OnStreamClosed is called immediately prior to closing an xDS stream.
func (m *xdsCallbackManager) OnStreamClosed(id int64, node *core.Node) {
	m.requestResponseLogger.OnStreamClosed(id, node)
}

// OnStreamRequest logs incoming SotW DiscoveryRequests.
func (m *xdsCallbackManager) OnStreamRequest(streamID int64, req *discoveryservice.DiscoveryRequest) error {
	return m.requestResponseLogger.OnStreamRequest(streamID, req)
}

// OnStreamResponse logs outgoing SotW DiscoveryResponses.
func (m *xdsCallbackManager) OnStreamResponse(ctx context.Context, streamID int64, req *discoveryservice.DiscoveryRequest, resp *discoveryservice.DiscoveryResponse) {
	m.requestResponseLogger.OnStreamResponse(ctx, streamID, req, resp)
}

// OnDeltaStreamOpen is called once an xDS Delta stream is open.
func (m *xdsCallbackManager) OnDeltaStreamOpen(ctx context.Context, id int64, typ string) error {
	return m.requestResponseLogger.OnDeltaStreamOpen(ctx, id, typ)
}

// OnDeltaStreamClosed is called immediately prior to closing an xDS Delta stream.
func (m *xdsCallbackManager) OnDeltaStreamClosed(id int64, node *core.Node) {
	m.requestResponseLogger.OnDeltaStreamClosed(id, node)
}

// OnStreamDeltaRequest logs incoming DeltaDiscoveryRequests and processes node metadata
func (m *xdsCallbackManager) OnStreamDeltaRequest(streamID int64, req *discoveryservice.DeltaDiscoveryRequest) error {
	// First call the logger's method
	if err := m.requestResponseLogger.OnStreamDeltaRequest(streamID, req); err != nil {
		return err
	}

	// Process node metadata if present
	if req.GetNode() != nil {
		nodeID := req.GetNode().GetId()
		if metadata := req.GetNode().GetMetadata(); metadata != nil {
			if initialClusters := metadata.GetFields()["initial_clusters"]; initialClusters != nil {
				if clusters := initialClusters.GetStructValue().GetFields()["clusters"]; clusters != nil {
					clusterList := clusters.GetListValue().GetValues()
					newClusters := make(map[string]struct{}, len(clusterList))

					for _, cluster := range clusterList {
						newClusters[cluster.GetStringValue()] = struct{}{}
					}

					// Check if the cluster set has changed
					m.mu.RLock()
					existingClusters, exists := m.nodeClusterCache[nodeID]
					m.mu.RUnlock()

					if !exists || !mapsEqual(existingClusters, newClusters) {
						// Update the LinearCache only if the set has changed
						// This won't trigger unnecessary rebuilds because the node that made the request
						// has not set
						if err := m.clusterCache.UpdateWilcardResourcesForNode(nodeID, newClusters); err != nil {
							log.Printf("Error updating clusters for node %s: %v", nodeID, err)
						} else {
							// Update our cache with the new set
							m.mu.Lock()
							m.nodeClusterCache[nodeID] = newClusters
							m.mu.Unlock()
							log.Printf("Updated cluster set for node %s: %v", nodeID, newClusters)
						}
					}
				}
			}
		}
	}

	return nil
}

// mapsEqual compares two maps for equality
func mapsEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

// OnStreamDeltaResponse logs outgoing DeltaDiscoveryResponses.
func (m *xdsCallbackManager) OnStreamDeltaResponse(streamID int64, req *discoveryservice.DeltaDiscoveryRequest, resp *discoveryservice.DeltaDiscoveryResponse) {
	m.requestResponseLogger.OnStreamDeltaResponse(streamID, req, resp)
}

// OnFetchRequest logs incoming Fetch DiscoveryRequests.
func (m *xdsCallbackManager) OnFetchRequest(ctx context.Context, req *discoveryservice.DiscoveryRequest) error {
	return m.requestResponseLogger.OnFetchRequest(ctx, req)
}

// OnFetchResponse logs outgoing Fetch DiscoveryResponses.
func (m *xdsCallbackManager) OnFetchResponse(req *discoveryservice.DiscoveryRequest, resp *discoveryservice.DiscoveryResponse) {
	m.requestResponseLogger.OnFetchResponse(req, resp)
}

// --- Main Function ---

func main() {
	// Setup file logger for detailed debug messages
	debugFile, err := os.OpenFile(debugLogFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open debug log file %s: %v", debugLogFilename, err)
	}
	defer func() {
		if err := debugFile.Close(); err != nil {
			log.Printf("Failed to close debug file: %v", err)
		}
	}()
	debugFileLogger := log.New(debugFile, "", log.LstdFlags) // Using standard log flags, no prefix from logger itself

	log.Printf("Control plane starting. Detailed xDS debug logs will be written to %s", debugLogFilename)

	// --- Initialize Caches ---
	listenerCache := xdscache.NewLinearCache(ListenerType, xdscache.WithCustomWildCardMode(false))
	clusterCache := xdscache.NewLinearCache(ClusterType, xdscache.WithCustomWildCardMode(true))
	routeCache := xdscache.NewLinearCache(RouteType, xdscache.WithCustomWildCardMode(false))
	endpointCache := xdscache.NewLinearCache(EndpointType, xdscache.WithCustomWildCardMode(false))
	virtualHostCache := xdscache.NewLinearCache(VirtualHostType, xdscache.WithCustomWildCardMode(true))

	// --- Initialize Empty Caches ---
	// Resources will be populated via CLI commands

	log.Printf("Initialized empty Linear caches. Use CLI to populate resources.")

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
	logger := &requestResponseLogger{debugFileLogger: debugFileLogger}
	cb := NewXdsCallbackManager(logger, virtualHostCache, clusterCache)

	// Server for ADS and VHDS
	srv := serverv3.NewServer(ctx, muxCache, cb)

	// --- Start gRPC Server ---
	// Remove or keep gRPC interceptor based on needs. For now, removing it as we have xDS layer logging.
	grpcServer := grpc.NewServer()
	// Listen on loopback only
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", gRPCport))
	if err != nil {
		log.Fatal(err)
	}

	// Register ADS server
	discoveryservice.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	// Register dedicated VHDS server
	routeservice.RegisterVirtualHostDiscoveryServiceServer(grpcServer, srv)
	// Register dedicated CDS server for ODCDS
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, srv)

	// Register ResourceManager service for CLI operations
	resourceManagerService := NewResourceManagerService(listenerCache, clusterCache, routeCache, endpointCache, virtualHostCache)
	pb.RegisterResourceManagerServer(grpcServer, resourceManagerService)

	log.Printf("xDS control plane: ADS, VHDS (Delta GRPC), and ResourceManager services listening on %d\n", gRPCport)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	// --- Start HTTP Server for Cache Display ---
	cacheHandler := NewCacheDisplayHandler(listenerCache, clusterCache, routeCache, endpointCache, virtualHostCache)
	http.Handle("/", cacheHandler)
	
	log.Printf("HTTP cache display server listening on http://localhost:%d\n", HTTPport)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", HTTPport), nil); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait indefinitely
	<-ctx.Done()
	grpcServer.GracefulStop()
}
