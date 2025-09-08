package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"controlplaneplayground/pb"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// ResourceManagerService implements the ResourceManager gRPC service
type ResourceManagerService struct {
	pb.UnimplementedResourceManagerServer
	listenerCache      *xdscache.LinearCache
	clusterCache       xdscache.SnapshotCache
	routeCache         *xdscache.LinearCache
	endpointCache      *xdscache.LinearCache
	virtualHostCache   *xdscache.LinearCache
	globalClusterStore map[string]*cluster.Cluster
	callbackManager    *xdsCallbackManager // Reference to update global store
	mu                 sync.RWMutex
}

// NewResourceManagerService creates a new ResourceManagerService instance
func NewResourceManagerService(listenerCache *xdscache.LinearCache, clusterCache xdscache.SnapshotCache, routeCache, endpointCache, virtualHostCache *xdscache.LinearCache, callbackManager *xdsCallbackManager) *ResourceManagerService {
	return &ResourceManagerService{
		listenerCache:      listenerCache,
		clusterCache:       clusterCache,
		routeCache:         routeCache,
		endpointCache:      endpointCache,
		virtualHostCache:   virtualHostCache,
		globalClusterStore: make(map[string]*cluster.Cluster),
		callbackManager:    callbackManager,
	}
}

// UpdateResource handles resource update/create requests
func (s *ResourceManagerService) UpdateResource(ctx context.Context, req *pb.UpdateResourceRequest) (*pb.UpdateResourceResponse, error) {
	// Log first 200 chars of data for debugging
	dataStr := string(req.Data)
	if len(dataStr) > 200 {
		dataStr = dataStr[:200] + "..."
	}

	// Handle different resource types
	var resource proto.Message
	var err error

	switch req.TypeUrl {
	case resourcev3.ListenerType:
		resource, err = s.deserializeListener(req.Data)
		if err == nil {
			err = s.listenerCache.UpdateResource(req.Name, resource)
		}
	case resourcev3.ClusterType:
		clusterResource, deserErr := s.deserializeCluster(req.Data)
		if deserErr != nil {
			err = deserErr
		} else {
			// Store in global cluster store and notify callback manager
			s.mu.Lock()
			s.globalClusterStore[req.Name] = clusterResource
			s.mu.Unlock()

			// Update callback manager's global store
			if s.callbackManager != nil {
				s.callbackManager.UpdateGlobalCluster(req.Name, clusterResource)
			}

			// Cluster added to global store (logged in summary)
		}
		resource = clusterResource
	case resourcev3.RouteType:
		resource, err = s.deserializeRoute(req.Data)
		if err == nil {
			err = s.routeCache.UpdateResource(req.Name, resource)
		}
	case resourcev3.EndpointType:
		resource, err = s.deserializeEndpoint(req.Data)
		if err == nil {
			err = s.endpointCache.UpdateResource(req.Name, resource)
		}
	case resourcev3.VirtualHostType:
		resource, err = s.deserializeVirtualHost(req.Data)
		if err == nil {
			err = s.virtualHostCache.UpdateResource(req.Name, resource)
		}
	default:
		log.Printf("DEBUG: Unsupported resource type: %s", req.TypeUrl)
		return &pb.UpdateResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Unsupported resource type: %s", req.TypeUrl),
		}, nil
	}

	if err != nil {
		return &pb.UpdateResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to process resource: %v", err),
		}, nil
	}

	// Log cache state for debugging
	s.logCacheState()

	return &pb.UpdateResourceResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully updated resource %s in cache", req.Name),
	}, nil
}

// DeleteResource handles resource deletion requests
func (s *ResourceManagerService) DeleteResource(ctx context.Context, req *pb.DeleteResourceRequest) (*pb.DeleteResourceResponse, error) {
	// Handle deletion based on resource type
	var err error
	switch req.TypeUrl {
	case resourcev3.ListenerType:
		err = s.listenerCache.DeleteResource(req.Name)
	case resourcev3.ClusterType:
		// Remove from global cluster store
		s.mu.Lock()
		delete(s.globalClusterStore, req.Name)
		s.mu.Unlock()
		// Cluster removed from global store
	case resourcev3.RouteType:
		err = s.routeCache.DeleteResource(req.Name)
	case resourcev3.EndpointType:
		err = s.endpointCache.DeleteResource(req.Name)
	case resourcev3.VirtualHostType:
		err = s.virtualHostCache.DeleteResource(req.Name)
	default:
		return &pb.DeleteResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Unsupported resource type: %s", req.TypeUrl),
		}, nil
	}

	if err != nil {
		return &pb.DeleteResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete resource: %v", err),
		}, nil
	}

	return &pb.DeleteResourceResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted resource %s from cache", req.Name),
	}, nil
}

// Deserialization functions using protojson unmarshaling
func (s *ResourceManagerService) deserializeListener(data []byte) (*listener.Listener, error) {
	unmarshaler := protojson.UnmarshalOptions{}
	var l listener.Listener
	if err := unmarshaler.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	return &l, nil
}

func (s *ResourceManagerService) deserializeCluster(data []byte) (*cluster.Cluster, error) {
	unmarshaler := protojson.UnmarshalOptions{}
	var c cluster.Cluster
	if err := unmarshaler.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	return &c, nil
}

func (s *ResourceManagerService) deserializeRoute(data []byte) (*route.RouteConfiguration, error) {
	unmarshaler := protojson.UnmarshalOptions{}
	var rc route.RouteConfiguration
	if err := unmarshaler.Unmarshal(data, &rc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	return &rc, nil
}

func (s *ResourceManagerService) deserializeEndpoint(data []byte) (*endpoint.ClusterLoadAssignment, error) {
	unmarshaler := protojson.UnmarshalOptions{}
	var cla endpoint.ClusterLoadAssignment
	if err := unmarshaler.Unmarshal(data, &cla); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	return &cla, nil
}

func (s *ResourceManagerService) deserializeVirtualHost(data []byte) (*route.VirtualHost, error) {
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true,
		Resolver:       protoregistry.GlobalTypes,
		AllowPartial:   true,
	}
	var vh route.VirtualHost
	if err := unmarshaler.Unmarshal(data, &vh); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	return &vh, nil
}

// logCacheState logs a summary of cache contents
func (s *ResourceManagerService) logCacheState() {
	// Get resources from LinearCaches
	listenerResources := s.listenerCache.GetResources()
	routeResources := s.routeCache.GetResources()
	endpointResources := s.endpointCache.GetResources()
	vhostResources := s.virtualHostCache.GetResources()

	// Log global cluster store
	s.mu.RLock()
	globalClusterCount := len(s.globalClusterStore)
	s.mu.RUnlock()

	// Log cluster snapshots
	snapshotNodeCount := len(s.clusterCache.GetStatusKeys())

	log.Printf("Cache summary: Listeners=%d, Routes=%d, Endpoints=%d, VirtualHosts=%d, GlobalClusters=%d, SnapshotNodes=%d",
		len(listenerResources), len(routeResources), len(endpointResources), len(vhostResources), globalClusterCount, snapshotNodeCount)
}
