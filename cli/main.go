package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"controlplaneplayground/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ResourceConfig represents a resource configuration from JSON
type ResourceConfig struct {
	TypeURL            string          `json:"type_url"`
	Name               string          `json:"name"`
	Data               json.RawMessage `json:"data"`                           // Use RawMessage to preserve JSON structure
	WildcardNodeUpdate bool            `json:"wildcard_node_update,omitempty"` // If true, call UpdateWildcardResourcesForNode
}

// Config represents the overall CLI configuration
type Config struct {
	ServerAddress string           `json:"server_address"`
	NodeID        string           `json:"node_id,omitempty"` // Default node ID for wildcard updates
	Resources     []ResourceConfig `json:"resources"`
}

func main() {
	var (
		configFile = flag.String("config", "", "Path to JSON configuration file")
		action     = flag.String("action", "update", "Action to perform: update or delete")
		serverAddr = flag.String("server", "localhost:18000", "gRPC server address")
	)
	flag.Parse()

	if *configFile == "" {
		log.Fatal("Please provide a configuration file with -config flag")
	}

	// Read and parse JSON file
	config, err := readConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// Override server address if provided via flag
	if *serverAddr != "localhost:18000" {
		config.ServerAddress = *serverAddr
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(config.ServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewResourceManagerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Process each resource based on action
	log.Printf("Processing %d resources...", len(config.Resources))
	for i, resource := range config.Resources {
		// Processing resource

		switch *action {
		case "update":
			err = updateResource(ctx, client, resource, config)
		case "delete":
			err = deleteResource(ctx, client, resource)
		default:
			log.Fatalf("Invalid action: %s. Use 'update' or 'delete'", *action)
		}

		if err != nil {
			log.Printf("ERROR: Failed to %s resource %s: %v", *action, resource.Name, err)
		} else {
			log.Printf("SUCCESS: %sed resource: %s (type: %s)", *action, resource.Name, resource.TypeURL)
		}
	}
}

func readConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Set default server address if not specified
	if config.ServerAddress == "" {
		config.ServerAddress = "localhost:18000"
	}

	return &config, nil
}

func updateResource(ctx context.Context, client pb.ResourceManagerClient, resource ResourceConfig, config *Config) error {
	// Convert JSON data to bytes
	dataBytes, err := json.Marshal(resource.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON data: %v", err)
	}

	// Sending update request

	// Request data prepared

	// Default node ID
	nodeID := "test-id"
	if config != nil && config.NodeID != "" {
		nodeID = config.NodeID
	}

	req := &pb.UpdateResourceRequest{
		TypeUrl:            resource.TypeURL,
		Name:               resource.Name,
		Data:               dataBytes, // Send JSON data as bytes
		WildcardNodeUpdate: resource.WildcardNodeUpdate,
		NodeId:             nodeID,
	}

	if resource.WildcardNodeUpdate {
		log.Printf("DEBUG: Will call UpdateWildcardResourcesForNode with nodeID: %s", nodeID)
	}

	resp, err := client.UpdateResource(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %v", err)
	}

	// Response received

	if !resp.Success {
		return fmt.Errorf("server returned error: %s", resp.Message)
	}

	fmt.Printf("Update response: %s\n", resp.Message)
	return nil
}

func deleteResource(ctx context.Context, client pb.ResourceManagerClient, resource ResourceConfig) error {
	req := &pb.DeleteResourceRequest{
		TypeUrl: resource.TypeURL,
		Name:    resource.Name,
	}

	resp, err := client.DeleteResource(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("server returned error: %s", resp.Message)
	}

	fmt.Printf("Delete response: %s\n", resp.Message)
	return nil
}
