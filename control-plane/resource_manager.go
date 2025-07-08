package main

import (
	"context"
	"fmt"
	"log"

	"controlplaneplayground/pb"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

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
	listenerCache    *xdscache.LinearCache
	clusterCache     *xdscache.LinearCache
	routeCache       *xdscache.LinearCache
	endpointCache    *xdscache.LinearCache
	virtualHostCache *xdscache.LinearCache
}

// NewResourceManagerService creates a new ResourceManagerService instance
func NewResourceManagerService(listenerCache, clusterCache, routeCache, endpointCache, virtualHostCache *xdscache.LinearCache) *ResourceManagerService {
	return &ResourceManagerService{
		listenerCache:    listenerCache,
		clusterCache:     clusterCache,
		routeCache:       routeCache,
		endpointCache:    endpointCache,
		virtualHostCache: virtualHostCache,
	}
}

// UpdateResource handles resource update/create requests
func (s *ResourceManagerService) UpdateResource(ctx context.Context, req *pb.UpdateResourceRequest) (*pb.UpdateResourceResponse, error) {
	log.Printf("UpdateResource called - TypeURL: %s, Name: %s", req.TypeUrl, req.Name)

	// Determine which cache to use based on TypeURL
	var targetCache *xdscache.LinearCache
	var resource proto.Message
	var err error

	switch req.TypeUrl {
	case resourcev3.ListenerType:
		targetCache = s.listenerCache
		resource, err = s.deserializeListener(req.Data)
	case resourcev3.ClusterType:
		targetCache = s.clusterCache
		resource, err = s.deserializeCluster(req.Data)
	case resourcev3.RouteType:
		targetCache = s.routeCache
		resource, err = s.deserializeRoute(req.Data)
	case resourcev3.EndpointType:
		targetCache = s.endpointCache
		resource, err = s.deserializeEndpoint(req.Data)
	case resourcev3.VirtualHostType:
		targetCache = s.virtualHostCache
		resource, err = s.deserializeVirtualHost(req.Data)
	default:
		return &pb.UpdateResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Unsupported resource type: %s", req.TypeUrl),
		}, nil
	}

	if err != nil {
		return &pb.UpdateResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to deserialize resource: %v", err),
		}, nil
	}

	// Update the cache
	if err := targetCache.UpdateResource(req.Name, resource); err != nil {
		return &pb.UpdateResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update cache: %v", err),
		}, nil
	}

	return &pb.UpdateResourceResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully updated resource %s in cache", req.Name),
	}, nil
}

// DeleteResource handles resource deletion requests
func (s *ResourceManagerService) DeleteResource(ctx context.Context, req *pb.DeleteResourceRequest) (*pb.DeleteResourceResponse, error) {
	log.Printf("DeleteResource called - TypeURL: %s, Name: %s", req.TypeUrl, req.Name)

	// Determine which cache to use based on TypeURL
	var targetCache *xdscache.LinearCache

	switch req.TypeUrl {
	case resourcev3.ListenerType:
		targetCache = s.listenerCache
	case resourcev3.ClusterType:
		targetCache = s.clusterCache
	case resourcev3.RouteType:
		targetCache = s.routeCache
	case resourcev3.EndpointType:
		targetCache = s.endpointCache
	case resourcev3.VirtualHostType:
		targetCache = s.virtualHostCache
	default:
		return &pb.DeleteResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Unsupported resource type: %s", req.TypeUrl),
		}, nil
	}

	// Delete from the cache
	if err := targetCache.DeleteResource(req.Name); err != nil {
		return &pb.DeleteResourceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to delete from cache: %v", err),
		}, nil
	}

	return &pb.DeleteResourceResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted resource %s from cache", req.Name),
	}, nil
}

// Deserialization functions using protojson unmarshaling
func (s *ResourceManagerService) deserializeListener(data []byte) (*listener.Listener, error) {
	unmarshaler := protojson.UnmarshalOptions{
	}
	var l listener.Listener
	if err := unmarshaler.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	// Debug logging to show populated fields
	log.Printf("Deserialized listener: Name=%s", l.Name)

	return &l, nil
}

func (s *ResourceManagerService) deserializeCluster(data []byte) (*cluster.Cluster, error) {
	unmarshaler := protojson.UnmarshalOptions{
	}
	var c cluster.Cluster
	if err := unmarshaler.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	// Debug logging to show populated fields
	log.Printf("Deserialized cluster: Name=%s, Type=%v, LbPolicy=%v, ConnectTimeout=%v",
		c.Name, c.GetType(), c.GetLbPolicy(), c.GetConnectTimeout())

	return &c, nil
}

func (s *ResourceManagerService) deserializeRoute(data []byte) (*route.RouteConfiguration, error) {
	unmarshaler := protojson.UnmarshalOptions{
	}
	var rc route.RouteConfiguration
	if err := unmarshaler.Unmarshal(data, &rc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	// Debug logging to show populated fields
	log.Printf("Deserialized route: Name=%s", rc.Name)

	return &rc, nil
}

func (s *ResourceManagerService) deserializeEndpoint(data []byte) (*endpoint.ClusterLoadAssignment, error) {
	unmarshaler := protojson.UnmarshalOptions{
	}
	var cla endpoint.ClusterLoadAssignment
	if err := unmarshaler.Unmarshal(data, &cla); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	// Debug logging to show populated fields
	log.Printf("Deserialized endpoint: ClusterName=%s, Endpoints=%d",
		cla.ClusterName, len(cla.Endpoints))
	if len(cla.Endpoints) > 0 && len(cla.Endpoints[0].LbEndpoints) > 0 {
		if addr := cla.Endpoints[0].LbEndpoints[0].GetEndpoint().GetAddress().GetSocketAddress(); addr != nil {
			log.Printf("  Endpoint address: %s:%d", addr.Address, addr.GetPortValue())
		}
	}

	return &cla, nil
}

func (s *ResourceManagerService) deserializeVirtualHost(data []byte) (*route.VirtualHost, error) {
	unmarshaler := protojson.UnmarshalOptions{
	}
	var vh route.VirtualHost
	if err := unmarshaler.Unmarshal(data, &vh); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to protobuf: %v", err)
	}

	// Debug logging to show populated fields
	log.Printf("Deserialized virtual host: Name=%s, Domains=%v", vh.Name, vh.Domains)

	return &vh, nil
}
