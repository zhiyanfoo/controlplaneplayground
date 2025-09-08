package main

import (
	"context"
	"controlplaneplayground/pb"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/on_demand/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	_ "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	debugLogFilename = "xds_debug.log"
	gRPCport         = 18000
	HTTPport         = 8734 // HTTP server for cache display

	// Resource Type URLs
	ListenerType    = resourcev3.ListenerType
	RouteType       = resourcev3.RouteType
	ClusterType     = resourcev3.ClusterType
	EndpointType    = resourcev3.EndpointType
	VirtualHostType = resourcev3.VirtualHostType
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

// CacheDisplayHandler handles HTTP requests to display cache contents
type CacheDisplayHandler struct {
	listenerCache    *xdscache.LinearCache
	clusterCache     xdscache.SnapshotCache
	routeCache       *xdscache.LinearCache
	endpointCache    *xdscache.LinearCache
	virtualHostCache *xdscache.LinearCache
}

// NewCacheDisplayHandler creates a new cache display handler
func NewCacheDisplayHandler(listenerCache *xdscache.LinearCache, clusterCache xdscache.SnapshotCache, routeCache, endpointCache, virtualHostCache *xdscache.LinearCache) *CacheDisplayHandler {
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
	routeResources := h.routeCache.GetResources()
	endpointResources := h.endpointCache.GetResources()
	vhostResources := h.virtualHostCache.GetResources()

	// Get clusters from all node snapshots
	clusterResources := make(map[string]proto.Message)
	for _, nodeID := range h.clusterCache.GetStatusKeys() {
		if snapshot, err := h.clusterCache.GetSnapshot(nodeID); err == nil {
			clusters := snapshot.GetResources(resourcev3.ClusterType)
			for name, resource := range clusters {
				clusterResources[name] = resource
			}
		}
	}

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
	virtualHostCache   *xdscache.LinearCache
	clusterCache       xdscache.SnapshotCache
	globalClusterStore map[string]*cluster.Cluster // Maps cluster name to cluster resource
	mu                 sync.RWMutex                // Protects globalClusterStore
}

// ClusterList holds the list of clusters for a node
type ClusterList struct {
	Clusters []string
}

// NewXdsCallbackManager creates a new xdsCallbackManager instance
func NewXdsCallbackManager(logger *requestResponseLogger, vhCache *xdscache.LinearCache, clCache xdscache.SnapshotCache) *xdsCallbackManager {
	return &xdsCallbackManager{
		requestResponseLogger: logger,
		virtualHostCache:      vhCache,
		clusterCache:          clCache,
		globalClusterStore:    make(map[string]*cluster.Cluster),
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

// OnStreamDeltaRequest logs incoming DeltaDiscoveryRequests and processes cluster requests
func (m *xdsCallbackManager) OnStreamDeltaRequest(streamID int64, req *discoveryservice.DeltaDiscoveryRequest) error {
	// First call the logger's method
	if err := m.requestResponseLogger.OnStreamDeltaRequest(streamID, req); err != nil {
		return err
	}

	// Handle cluster requests specifically
	if req.GetTypeUrl() == resourcev3.ClusterType && req.GetNode() != nil {
		nodeID := req.GetNode().GetId()
		requestedClusters := req.GetResourceNamesSubscribe()

		// Check if this is NOT a wildcard-only request
		if !isWildcardOnlyRequest(requestedClusters) {
			// Extract specific cluster names (ignore any "*" entries)
			specificClusters := extractSpecificClusters(requestedClusters)

			if len(specificClusters) > 0 {
				// This is an on-demand request - update node's snapshot
				m.mu.Lock()
				err := createOrUpdateNodeSnapshot(nodeID, specificClusters, m.clusterCache, m.globalClusterStore)
				m.mu.Unlock()

				if err != nil {
					log.Printf("Error creating snapshot for node %s: %v", nodeID, err)
				} else {
					log.Printf("Created snapshot for node %s with clusters: %v", nodeID, specificClusters)
				}
			}
		}
	}

	return nil
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

// UpdateGlobalCluster adds or updates a cluster in the global store
func (m *xdsCallbackManager) UpdateGlobalCluster(name string, cluster *cluster.Cluster) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.globalClusterStore[name] = cluster
	log.Printf("Updated global cluster store with cluster: %s", name)
}

// --- Helper Functions ---

// generateSnapshotVersion creates a time-based version string
func generateSnapshotVersion() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// isWildcardOnlyRequest checks if the request is for wildcard resources only
func isWildcardOnlyRequest(resourceNames []string) bool {
	// Empty request or single "*" = wildcard only
	if len(resourceNames) == 0 {
		return true
	}
	if len(resourceNames) == 1 && resourceNames[0] == "*" {
		return true
	}
	return false
}

// extractSpecificClusters returns cluster names, filtering out wildcards
func extractSpecificClusters(resourceNames []string) []string {
	var specificClusters []string
	for _, name := range resourceNames {
		if name != "*" {
			specificClusters = append(specificClusters, name)
		}
	}
	return specificClusters
}

// createOrUpdateNodeSnapshot creates or updates a snapshot for a specific node
func createOrUpdateNodeSnapshot(nodeID string, requestedClusters []string, snapshotCache xdscache.SnapshotCache, globalStore map[string]*cluster.Cluster) error {
	// Collect requested clusters that exist in global store
	var clusters []types.Resource
	for _, clusterName := range requestedClusters {
		if cluster, exists := globalStore[clusterName]; exists {
			clusters = append(clusters, cluster)
		}
	}

	if len(clusters) == 0 {
		return nil // No valid clusters to add
	}

	// Create new snapshot
	version := generateSnapshotVersion()
	resources := map[resourcev3.Type][]types.Resource{
		resourcev3.ClusterType: clusters,
	}

	snapshot, err := xdscache.NewSnapshot(version, resources)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}

	return snapshotCache.SetSnapshot(context.Background(), nodeID, snapshot)
}

// --- Main Function ---

func main() {
	// Parse command line flags
	var bindAll bool
	flag.BoolVar(&bindAll, "bind-all", false, "Bind to 0.0.0.0 instead of 127.0.0.1 (required for Docker)")
	flag.Parse()

	// Determine bind address
	bindAddr := "127.0.0.1"
	if bindAll {
		bindAddr = "0.0.0.0"
	}

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
	clusterCache := xdscache.NewSnapshotCache(false, xdscache.IDHash{}, nil)
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

	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindAddr, gRPCport))
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
	resourceManagerService := NewResourceManagerService(listenerCache, clusterCache, routeCache, endpointCache, virtualHostCache, cb)
	pb.RegisterResourceManagerServer(grpcServer, resourceManagerService)

	log.Printf("xDS control plane listening on %s:%d\n", bindAddr, gRPCport)
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
