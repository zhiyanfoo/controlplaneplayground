package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"controlplaneplayground/testpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Connect to Envoy proxy
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := testpb.NewTestServiceClient(conn)

	// Test without session IDs first to let envoy generate them
	var sessionID string
	fmt.Printf("\n=== Testing without session ID (envoy should generate) ===\n")
	for i := 0; i < 10; i++ {
		ctx := context.Background()
		var header metadata.MD
		
		// Use the session ID from previous request if we have one
		if sessionID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-session-id", sessionID)
		}
		
		req := &testpb.HelloRequest{
			Name: fmt.Sprintf("User-%d", i+1),
		}
		
		resp, err := client.SayHello(ctx, req, grpc.Header(&header))
		if err != nil {
			log.Printf("Request failed: %v", err)
			continue
		}
		
		if sessionHeaders := header.Get("x-session-id"); len(sessionHeaders) > 0 {
			sessionID = sessionHeaders[0]
		}
		
		fmt.Printf("Request %d [session: %s]: %s\n", i+1, sessionID, resp.Message)
		time.Sleep(100 * time.Millisecond)
	}
}
